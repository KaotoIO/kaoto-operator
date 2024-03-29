package designer

import (
	"context"

	"github.com/kaotoIO/kaoto-operator/pkg/controller/client"

	"github.com/kaotoIO/kaoto-operator/pkg/apply"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewServiceAction() Action {
	return &serviceAction{}
}

type serviceAction struct {
}

func (a *serviceAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	b = b.Owns(&corev1.Service{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))

	return b, nil
}

func (a *serviceAction) Cleanup(context.Context, *ReconciliationRequest) error {
	return nil
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
