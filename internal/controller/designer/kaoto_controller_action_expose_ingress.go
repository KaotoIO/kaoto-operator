package designer

import (
	"context"
	"fmt"
	"strings"

	"github.com/kaotoIO/kaoto-operator/pkg/controller/client"

	"github.com/kaotoIO/kaoto-operator/pkg/apply"

	"github.com/kaotoIO/kaoto-operator/pkg/controller/predicates"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	netv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	netv1ac "k8s.io/client-go/applyconfigurations/networking/v1"
)

func NewIngressAction() Action {
	return &ingressAction{}
}

type ingressAction struct {
}

func (a *ingressAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	b = b.Owns(&netv1.Ingress{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
			predicates.StatusChanged{},
		)))

	return b, nil
}

func (a *ingressAction) Cleanup(context.Context, *ReconciliationRequest) error {
	return nil
}

func (a *ingressAction) Apply(ctx context.Context, rr *ReconciliationRequest) error {
	ingressCondition := metav1.Condition{
		Type:               "Ingress",
		Status:             metav1.ConditionTrue,
		Reason:             "Deployed",
		Message:            "Deployed",
		ObservedGeneration: rr.Kaoto.Generation,
	}

	if rr.Kaoto.Spec.Ingress != nil {

		if err := a.ingress(ctx, rr); err != nil {
			ingressCondition.Status = metav1.ConditionFalse
			ingressCondition.Reason = "Failure"
			ingressCondition.Message = err.Error()

			return err
		}

	} else {
		ingressCondition.Status = metav1.ConditionFalse
		ingressCondition.Reason = "NotRequires"
		ingressCondition.Message = "NotRequires"

		if err := a.cleanup(ctx, rr); err != nil {
			ingressCondition.Status = metav1.ConditionFalse
			ingressCondition.Reason = "Failure"
			ingressCondition.Message = err.Error()

			return err
		}
	}

	var in netv1.Ingress

	if err := rr.Get(ctx, rr.Key(), &in); err != nil && !k8serrors.IsNotFound(err) {
		ingressCondition.Status = metav1.ConditionFalse
		ingressCondition.Reason = "Failure"
		ingressCondition.Message = err.Error()
	} else {
		rr.Kaoto.Status.Endpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local/", rr.Kaoto.Name, rr.Kaoto.Namespace)

		if len(in.Status.LoadBalancer.Ingress) > 0 {
			switch {
			case in.Status.LoadBalancer.Ingress[0].Hostname != "":
				rr.Kaoto.Status.Endpoint = "http://" + in.Status.LoadBalancer.Ingress[0].Hostname + "/" + rr.Kaoto.Name
			case in.Status.LoadBalancer.Ingress[0].IP != "":
				rr.Kaoto.Status.Endpoint = "http://" + in.Status.LoadBalancer.Ingress[0].IP + "/" + rr.Kaoto.Name
			}
		}

		if !strings.HasSuffix(rr.Kaoto.Status.Endpoint, "/") {
			rr.Kaoto.Status.Endpoint += "/"
		}
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, ingressCondition)

	return nil
}

func (a *ingressAction) ingress(ctx context.Context, rr *ReconciliationRequest) error {
	host := ""
	path := "/" + rr.Kaoto.Name + "(/|$)(.*)"

	if rr.Kaoto.Spec.Ingress.Host != "" {
		host = rr.Kaoto.Spec.Ingress.Host
	}
	if rr.Kaoto.Spec.Ingress.Path != "" {
		path = rr.Kaoto.Spec.Ingress.Path
	}

	if !strings.HasSuffix(path, "(/|$)(.*)") {
		path += "(/|$)(.*)"
	}

	resource := netv1ac.Ingress(rr.Kaoto.Name, rr.Kaoto.Namespace).
		WithOwnerReferences(apply.WithOwnerReference(rr.Kaoto)).
		WithAnnotations(map[string]string{
			"nginx.ingress.kubernetes.io/use-regex":      "true",
			"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
		}).
		WithLabels(Labels(rr.Kaoto)).
		WithSpec(netv1ac.IngressSpec().
			WithRules(netv1ac.IngressRule().
				WithHost(host).
				WithHTTP(netv1ac.HTTPIngressRuleValue().
					WithPaths(netv1ac.HTTPIngressPath().
						WithPath(path).
						WithPathType(netv1.PathTypeImplementationSpecific).
						WithBackend(netv1ac.IngressBackend().
							WithService(netv1ac.IngressServiceBackend().
								WithName(rr.Kaoto.Name).
								WithPort(netv1ac.ServiceBackendPort().
									WithName(KaotoPortType))))))))

	_, err := rr.Client.NetworkingV1().Ingresses(rr.Kaoto.Namespace).Apply(
		ctx,
		resource,
		metav1.ApplyOptions{
			FieldManager: KaotoOperatorFieldManager,
			Force:        true,
		},
	)

	return err
}

func (a *ingressAction) cleanup(ctx context.Context, rr *ReconciliationRequest) error {
	ingress := netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rr.Kaoto.Name,
			Namespace: rr.Kaoto.Namespace,
		},
	}

	if err := rr.Client.Delete(ctx, &ingress); err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}
