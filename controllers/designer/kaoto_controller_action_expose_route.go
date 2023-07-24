package designer

import (
	"context"
	"fmt"
	"strings"

	"github.com/kaotoIO/kaoto-operator/config/apply"

	routev1 "github.com/openshift/api/route/v1"
	routev1ac "github.com/openshift/client-go/route/applyconfigurations/route/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var routeRewriteAnnotations = map[string]string{
	"haproxy.router.openshift.io/rewrite-target": "/",
}

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

	var in routev1.Route

	if err := rr.Get(ctx, rr.NamespacedName, &in); err != nil && !k8serrors.IsNotFound(err) {
		ingressCondition.Status = metav1.ConditionFalse
		ingressCondition.Reason = "Failure"
		ingressCondition.Message = err.Error()
	} else {
		rr.Kaoto.Status.Endpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local/", rr.Kaoto.Name, rr.Kaoto.Namespace)

		if len(in.Status.Ingress) > 0 {
			switch {
			case in.Status.Ingress[0].Host != "":
				rr.Kaoto.Status.Endpoint = "https://" + in.Status.Ingress[0].Host + in.Spec.Path
			}
		}

		if !strings.HasSuffix(rr.Kaoto.Status.Endpoint, "/") {
			rr.Kaoto.Status.Endpoint = rr.Kaoto.Status.Endpoint + "/"
		}
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, ingressCondition)

	return nil
}

func (a *routeAction) route(ctx context.Context, rr *ReconciliationRequest) error {
	host := ""
	path := "/"

	if rr.Kaoto.Spec.Ingress.Host != "" {
		host = rr.Kaoto.Spec.Ingress.Host
	}
	if rr.Kaoto.Spec.Ingress.Path != "" {
		path = rr.Kaoto.Spec.Ingress.Path
	}

	resource := routev1ac.Route(rr.Kaoto.Name, rr.Kaoto.Namespace).
		WithOwnerReferences(apply.WithOwnerReference(rr.Kaoto)).
		WithAnnotations(a.rewriteAnnotations(rr)).
		WithSpec(routev1ac.RouteSpec().
			WithHost(host).
			WithPath(path).
			WithTo(routev1ac.RouteTargetReference().
				WithKind("Service").
				WithName(rr.Kaoto.Name)).
			WithPort(routev1ac.RoutePort().
				WithTargetPort(intstr.FromInt(int(KaotoPort)))).
			WithTLS(routev1ac.TLSConfig().
				WithTermination(routev1.TLSTerminationEdge).
				WithInsecureEdgeTerminationPolicy(routev1.InsecureEdgeTerminationPolicyRedirect)))

	_, err := rr.Route.RouteV1().Routes(rr.Kaoto.Namespace).Apply(
		ctx,
		resource,
		metav1.ApplyOptions{
			FieldManager: KaotoOperatorFieldManager,
			Force:        true,
		},
	)

	return err
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

func (a *routeAction) rewriteAnnotations(rr *ReconciliationRequest) map[string]string {
	if rr.Kaoto.Spec.Ingress.Path != "" {
		return routeRewriteAnnotations
	}

	return nil
}
