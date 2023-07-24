package designer

import (
	"context"
	"strings"

	"github.com/kaotoIO/kaoto-operator/pkg/defaults"

	routev1 "github.com/openshift/api/route/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kaotoIO/kaoto-operator/pkg/pointer"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type deployAction struct {
}

func (a *deployAction) Apply(ctx context.Context, rr *ReconciliationRequest) error {
	deploymentCondition := metav1.Condition{
		Type:               "Deployment",
		Status:             metav1.ConditionTrue,
		Reason:             "Deployed",
		Message:            "Deployed",
		ObservedGeneration: rr.Kaoto.Generation,
	}

	err := a.deploy(ctx, rr)
	if err != nil {
		deploymentCondition.Status = metav1.ConditionFalse
		deploymentCondition.Reason = "Failure"
		deploymentCondition.Message = err.Error()
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, deploymentCondition)

	return err
}

const (
	KaotoPort               int32  = 8081
	KaotoPortType           string = "http"
	KaotoLivenessProbePath  string = "/q/health/live"
	KaotoReadinessProbePath string = "/q/health/ready"
)

func (a *deployAction) deploy(ctx context.Context, rr *ReconciliationRequest) error {
	return reify(
		ctx,
		rr,
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name,
				Namespace: rr.Kaoto.Namespace,
			},
		},
		func(resource *appsv1.Deployment) (*appsv1.Deployment, error) {
			if err := controllerutil.SetControllerReference(rr.Kaoto, resource, rr.Scheme()); err != nil {
				return resource, errors.New("unable to set controller reference")
			}

			image := rr.Kaoto.Spec.Image
			if image == "" {
				image = defaults.KaotoStandaloneImage
			}

			resource.Spec = appsv1.DeploymentSpec{
				Replicas: pointer.Any(int32(1)),
				Selector: &metav1.LabelSelector{
					MatchLabels: LabelsForSelector(rr.Kaoto),
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: Labels(rr.Kaoto),
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: rr.Kaoto.Name,
						Containers: []corev1.Container{{
							Image: image,
							Name:  "standalone",
							Env: []corev1.EnvVar{{
								Name: "NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							}},
							Ports: []corev1.ContainerPort{{
								ContainerPort: KaotoPort,
								Name:          KaotoPortType,
							}},
							ImagePullPolicy: "Always",
							ReadinessProbe: &corev1.Probe{
								InitialDelaySeconds: 1,
								PeriodSeconds:       1,
								FailureThreshold:    3,
								SuccessThreshold:    1,
								TimeoutSeconds:      10,
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   KaotoReadinessProbePath,
										Port:   intstr.IntOrString{IntVal: KaotoPort},
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								InitialDelaySeconds: 1,
								PeriodSeconds:       1,
								FailureThreshold:    3,
								SuccessThreshold:    1,
								TimeoutSeconds:      10,
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   KaotoLivenessProbePath,
										Port:   intstr.IntOrString{IntVal: KaotoPort},
										Scheme: corev1.URISchemeHTTP,
									},
								},
							},
						}},
					},
				},
			}

			// it appears that even with the standalone mode, there are some CORS issues that prevent some endpoint
			// to be invoked, the QUARKUS_HTTP_CORS_ORIGINS env must be set.
			//
			// TODO: investigate why this is needed

			if rr.Kaoto.Spec.Ingress != nil {
				switch {
				case rr.Kaoto.Spec.Ingress.Host != "":
					// use the provided host if configured so no further lookup would be needed
					resource.Spec.Template.Spec.Containers[0].Env = append(resource.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
						Name:  "QUARKUS_HTTP_CORS_ORIGINS",
						Value: rr.Kaoto.Spec.Ingress.Host,
					})
				case rr.ClusterType == ClusterTypeOpenShift:
					// in case of OpenShift se can leverage the Route.Spec.Host to retrieve the
					// right value for the QUARKUS_HTTP_CORS_ORIGINS
					var in routev1.Route

					err := rr.Get(ctx, rr.NamespacedName, &in)
					if err != nil && !k8serrors.IsNotFound(err) {
						return resource, err
					}

					if in.Spec.Host != "" {
						host := in.Spec.Host
						if !strings.HasPrefix(host, "https://") {
							host = "https://" + host
						}
						resource.Spec.Template.Spec.Containers[0].Env = append(resource.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
							Name:  "QUARKUS_HTTP_CORS_ORIGINS",
							Value: host,
						})
					}
				}
			}

			return resource, nil
		},
	)
}
