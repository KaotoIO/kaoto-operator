package designer

import (
	"context"

	"github.com/kaotoIO/kaoto-operator/config/apply"

	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	rbacv1ac "k8s.io/client-go/applyconfigurations/rbac/v1"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type rbacAction struct {
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
	resource := rbacv1ac.ClusterRoleBinding(rr.Kaoto.Name).
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
