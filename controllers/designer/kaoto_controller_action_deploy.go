package designer

import (
	"context"

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
							Image: rr.Kaoto.Spec.Image,
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

			return resource, nil
		},
	)
}
