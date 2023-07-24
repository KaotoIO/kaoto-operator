package designer

import (
	"context"

	"github.com/kaotoIO/kaoto-operator/config/apply"

	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	service := corev1ac.Service(rr.Kaoto.Name, rr.Kaoto.Namespace).
		WithOwnerReferences(apply.WithOwnerReference(rr.Kaoto)).
		WithLabels(Labels(rr.Kaoto)).
		WithSpec(corev1ac.ServiceSpec().
			WithPorts(corev1ac.ServicePort().
				WithName(KaotoPortType).
				WithProtocol(corev1.ProtocolTCP).
				WithPort(KaotoPort).
				WithTargetPort(intstr.FromString(KaotoPortType))).
			WithSelector(LabelsForSelector(rr.Kaoto)).
			WithSessionAffinity(corev1.ServiceAffinityNone).
			WithPublishNotReadyAddresses(true))

	_, err := rr.Client.CoreV1().Services(rr.Kaoto.Namespace).Apply(
		ctx,
		service,
		metav1.ApplyOptions{
			FieldManager: KaotoOperatorFieldManager,
			Force:        true,
		},
	)

	return err
}
