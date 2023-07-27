package designer

import (
	"context"
	"fmt"

	"github.com/kaotoIO/kaoto-operator/pkg/apply"

	"github.com/kaotoIO/kaoto-operator/pkg/client"

	"github.com/kaotoIO/kaoto-operator/apis/designer/v1alpha1"

	"github.com/go-logr/logr"
	"github.com/kaotoIO/kaoto-operator/pkg/pointer"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	ctrlcl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	rbacv1ac "k8s.io/client-go/applyconfigurations/rbac/v1"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRBACAction() Action {
	return &rbacAction{
		l: ctrl.Log.WithName("action-rbac"),
	}
}

type rbacAction struct {
	l logr.Logger
}

func (a *rbacAction) Cleanup(ctx context.Context, rr *ReconciliationRequest) error {

	//
	// Some resources cannot be garbage collected and automatically removed using ownership as i.e. they
	// are cluster scoped such as a ClusterRoleBinding hence must be deleted
	//

	err := rr.Client.RbacV1().ClusterRoleBindings().Delete(ctx, rr.Kaoto.Namespace+"-"+rr.Kaoto.Name, metav1.DeleteOptions{
		PropagationPolicy: pointer.Any(metav1.DeletePropagationForeground),
	})

	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failure removing ClusterRoleBindings %s-%s", rr.Kaoto.Namespace, rr.Kaoto.Name)
	}

	return nil
}

func (a *rbacAction) Configure(_ context.Context, c *client.Client, b *builder.Builder) (*builder.Builder, error) {

	b = b.Owns(&corev1.ServiceAccount{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))

	b = b.Watches(
		&rbacv1.ClusterRoleBinding{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object ctrlcl.Object) []reconcile.Request {
			crb, ok := object.(*rbacv1.ClusterRoleBinding)
			if !ok {
				a.l.Error(fmt.Errorf("type assertion failed: %v", object), "failed to retrieve ClusterRoleBinding")
				return nil
			}

			if crb.Labels != nil &&
				crb.Labels[KubernetesLabelAppManagedBy] == KaotoOperatorFieldManager &&
				len(crb.Subjects) == 1 &&
				crb.Subjects[0].Kind == rbacv1.ServiceAccountKind {

				nn := types.NamespacedName{
					Name:      crb.Subjects[0].Name,
					Namespace: crb.Subjects[0].Namespace,
				}

				var kaoto v1alpha1.Kaoto
				err := c.Get(ctx, nn, &kaoto)
				if err != nil {
					if k8serrors.IsNotFound(err) {
						// no CR found, likely already deleted
						return nil
					}

					a.l.Error(err, "failed to retrieve Kaoto", "name", nn.Name, "namespace", nn.Namespace)

					return nil
				}

				if !kaoto.ObjectMeta.DeletionTimestamp.IsZero() {
					// object being deleting, nothing to do here
					return nil
				}

				//
				// The ClusterRoleBinding is defined as follows:
				//
				// bracv1ac.ClusterRoleBinding(rr.Kaoto.Namespace + "-" + rr.Kaoto.Name).
				//		WithSubjects(rbacv1ac.Subject().
				//			WithKind(rbacv1.ServiceAccountKind).
				//			WithNamespace(rr.Kaoto.Namespace).
				//			WithName(rr.Kaoto.Name))
				//
				// Hence we can use the subject's name and namespace to trigger a reconcile loop
				// to the related Kaoto resource
				//
				return []reconcile.Request{{
					NamespacedName: types.NamespacedName{
						Name:      crb.Subjects[0].Name,
						Namespace: crb.Subjects[0].Namespace,
					},
				}}
			}

			return nil

		}))

	return b, nil
}

func (a *rbacAction) Apply(ctx context.Context, rr *ReconciliationRequest) error {

	var allErrors error

	//
	// RBAC - ServiceAccount
	//

	saCondition := metav1.Condition{
		Type:               "ServiceAccount",
		Status:             metav1.ConditionTrue,
		Reason:             "Deployed",
		Message:            "Deployed",
		ObservedGeneration: rr.Kaoto.Generation,
	}

	if err := a.serviceAccount(ctx, rr); err != nil {
		saCondition.Status = metav1.ConditionFalse
		saCondition.Reason = "Failure"
		saCondition.Message = err.Error()

		allErrors = multierr.Append(allErrors, err)
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, saCondition)

	//
	// RBAC - RoleBinding
	//

	rbCondition := metav1.Condition{
		Type:               "ClusterRoleBinding",
		Status:             metav1.ConditionTrue,
		Reason:             "Deployed",
		Message:            "Deployed",
		ObservedGeneration: rr.Kaoto.Generation,
	}

	if err := a.binding(ctx, rr); err != nil {
		rbCondition.Status = metav1.ConditionFalse
		rbCondition.Reason = "Failure"
		rbCondition.Message = err.Error()

		allErrors = multierr.Append(allErrors, err)
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, rbCondition)

	return allErrors
}

func (a *rbacAction) serviceAccount(ctx context.Context, rr *ReconciliationRequest) error {
	resource := corev1ac.ServiceAccount(rr.Kaoto.Name, rr.Kaoto.Namespace).
		WithOwnerReferences(apply.WithOwnerReference(rr.Kaoto))

	_, err := rr.Client.CoreV1().ServiceAccounts(rr.Kaoto.Namespace).Apply(
		ctx,
		resource,
		metav1.ApplyOptions{
			FieldManager: KaotoOperatorFieldManager,
			Force:        true,
		},
	)

	return err
}

func (a *rbacAction) binding(ctx context.Context, rr *ReconciliationRequest) error {
	// A ClusterRoleBinding is not namespaced hence, the name must be made unique
	resource := rbacv1ac.ClusterRoleBinding(rr.Kaoto.Namespace + "-" + rr.Kaoto.Name).
		WithLabels(Labels(rr.Kaoto)).
		WithSubjects(rbacv1ac.Subject().
			WithKind(rbacv1.ServiceAccountKind).
			WithNamespace(rr.Kaoto.Namespace).
			WithName(rr.Kaoto.Name)).
		WithRoleRef(rbacv1ac.RoleRef().
			WithAPIGroup(rbacv1.GroupName).
			WithKind("ClusterRole").
			WithName(KaotoDeploymentClusterRoleName))

	_, err := rr.Client.RbacV1().ClusterRoleBindings().Apply(
		ctx,
		resource,
		metav1.ApplyOptions{
			FieldManager: KaotoOperatorFieldManager,
			Force:        true,
		},
	)

	return err
}
