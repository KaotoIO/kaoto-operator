package designer

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type serviceAction struct {
}

func (a *serviceAction) Apply(ctx context.Context, rr *ReconciliationRequest) error {
	serviceCondition := metav1.Condition{
		Type:               "Service",
		Status:             metav1.ConditionTrue,
		Reason:             "Deployed",
		Message:            "Deployed",
		ObservedGeneration: rr.Kaoto.Generation,
	}

	err := a.service(ctx, rr)
	if err != nil {
		serviceCondition.Status = metav1.ConditionFalse
		serviceCondition.Reason = "Failure"
		serviceCondition.Message = err.Error()

		return err
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, serviceCondition)

	return nil
}

func (a *serviceAction) service(ctx context.Context, rr *ReconciliationRequest) error {
	return reify(
		ctx,
		rr,
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name,
				Namespace: rr.Kaoto.Namespace,
			},
		},
		func(resource *corev1.Service) (*corev1.Service, error) {
			if err := controllerutil.SetControllerReference(rr.Kaoto, resource, rr.Scheme()); err != nil {
				return resource, errors.New("unable to set controller reference")
			}

			resource.Spec = corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.FromInt(8081),
					},
				},
				Selector:                 LabelsForSelector(rr.Kaoto),
				SessionAffinity:          "None",
				PublishNotReadyAddresses: true,
			}

			return resource, nil
		},
	)
}
