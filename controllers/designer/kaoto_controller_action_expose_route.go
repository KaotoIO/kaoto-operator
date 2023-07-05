package designer

import (
	"context"

	"github.com/kaotoIO/kaoto-operator/pkg/resources"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type routeAction struct {
}

func (a *routeAction) Apply(ctx context.Context, rr *ReconciliationRequest) error {
	ingressCondition := metav1.Condition{
		Type:               "Ingress",
		Status:             metav1.ConditionTrue,
		Reason:             "Deployed",
		Message:            "Deployed",
		ObservedGeneration: rr.Kaoto.Generation,
	}

	if rr.Kaoto.Spec.Ingress != nil {

		if err := a.route(ctx, rr); err != nil {
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

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, ingressCondition)

	return nil
}

func (a *routeAction) route(ctx context.Context, rr *ReconciliationRequest) error {
	return reify(
		ctx,
		rr,
		&routev1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name,
				Namespace: rr.Kaoto.Namespace,
			},
		},
		func(resource *routev1.Route) (*routev1.Route, error) {
			if err := controllerutil.SetControllerReference(rr.Kaoto, resource, rr.Scheme()); err != nil {
				return resource, errors.New("unable to set controller reference")
			}

			resources.SetAnnotation(resource, "haproxy.router.openshift.io/rewrite-target", "/")

			resource.Spec = routev1.RouteSpec{
				To: routev1.RouteTargetReference{
					Kind: "Service",
					Name: rr.Kaoto.Name,
				},
				Port: &routev1.RoutePort{
					TargetPort: intstr.FromInt(8081),
				},
				TLS: &routev1.TLSConfig{
					Termination:                   routev1.TLSTerminationEdge,
					InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
				},
			}

			host := ""
			path := "/" + rr.Kaoto.Name

			if rr.Kaoto.Spec.Ingress.Host != "" {
				host = rr.Kaoto.Spec.Ingress.Host
			}
			if rr.Kaoto.Spec.Ingress.Path != "" {
				path = rr.Kaoto.Spec.Ingress.Path
			}

			resource.Spec.Host = host
			resource.Spec.Path = path

			return resource, nil
		},
	)
}

func (a *routeAction) cleanup(ctx context.Context, rr *ReconciliationRequest) error {
	route := routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rr.Kaoto.Name,
			Namespace: rr.Kaoto.Namespace,
		},
	}

	if err := rr.Client.Delete(ctx, &route); err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	return nil
}
