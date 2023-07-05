package designer

import (
	"context"
	"strings"

	"github.com/kaotoIO/kaoto-operator/pkg/resources"

	"github.com/kaotoIO/kaoto-operator/pkg/pointer"
	"github.com/pkg/errors"
	netv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ingressAction struct {
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

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, ingressCondition)

	return nil
}

func (a *ingressAction) ingress(ctx context.Context, rr *ReconciliationRequest) error {
	return reify(
		ctx,
		rr,
		&netv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name,
				Namespace: rr.Kaoto.Namespace,
			},
		},
		func(resource *netv1.Ingress) (*netv1.Ingress, error) {
			if err := controllerutil.SetControllerReference(rr.Kaoto, resource, rr.Scheme()); err != nil {
				return resource, errors.New("unable to set controller reference")
			}

			resources.SetAnnotation(resource, "nginx.ingress.kubernetes.io/use-regex", "true")
			resources.SetAnnotation(resource, "nginx.ingress.kubernetes.io/rewrite-target", "/$2")

			host := ""
			path := "/" + rr.Kaoto.Name + "(/|$)(.*)"

			if rr.Kaoto.Spec.Ingress.Host != "" {
				host = rr.Kaoto.Spec.Ingress.Host
			}
			if rr.Kaoto.Spec.Ingress.Path != "" {
				path = rr.Kaoto.Spec.Ingress.Path
			}

			if !strings.HasSuffix(path, "(/|$)(.*)") {
				path = path + "(/|$)(.*)"
			}

			resource.Spec = netv1.IngressSpec{
				Rules: []netv1.IngressRule{
					{
						Host: host,
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Path:     path,
										PathType: pointer.Any(netv1.PathTypePrefix),
										Backend: netv1.IngressBackend{
											Service: &netv1.IngressServiceBackend{
												Name: rr.Kaoto.Name,
												Port: netv1.ServiceBackendPort{
													Name: "http",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			return resource, nil
		},
	)
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
