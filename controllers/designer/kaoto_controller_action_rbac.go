package designer

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	if err := reify(
		ctx,
		rr,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name,
				Namespace: rr.Kaoto.Namespace,
			},
		},
		func(resource *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
			if err := controllerutil.SetControllerReference(rr.Kaoto, resource, rr.Scheme()); err != nil {
				return resource, errors.New("unable to set controller reference")
			}

			return resource, nil
		},
	); err != nil {
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

	if err := reify(
		ctx,
		rr,
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name,
				Namespace: rr.Kaoto.Namespace,
			},
		},
		func(resource *rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
			resource.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "kaoto-backend",
			}
			resource.Subjects = []rbacv1.Subject{{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      rr.Kaoto.Name,
				Namespace: rr.Kaoto.Namespace,
			}}

			return resource, nil
		},
	); err != nil {
		rbCondition.Status = metav1.ConditionFalse
		rbCondition.Reason = "Failure"
		rbCondition.Message = err.Error()

		allErrors = multierr.Append(allErrors, err)
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, rbCondition)

	return allErrors
}
