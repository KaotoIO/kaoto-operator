package designer

import (
	"context"
	"strings"

	"github.com/kaotoIO/kaoto-operator/config/apply"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"

	"github.com/kaotoIO/kaoto-operator/pkg/defaults"

	routev1 "github.com/openshift/api/route/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
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

func (a *deployAction) deploy(ctx context.Context, rr *ReconciliationRequest) error {

	d, err := a.deployment(ctx, rr)
	if err != nil {
		return err
	}

	_, err = rr.Client.AppsV1().Deployments(rr.Kaoto.Namespace).Apply(
		ctx,
		d,
		metav1.ApplyOptions{
			FieldManager: KaotoOperatorFieldManager,
			Force:        true,
		},
	)

	return err
}

func (a *deployAction) deployment(ctx context.Context, rr *ReconciliationRequest) (*appsv1ac.DeploymentApplyConfiguration, error) {
	image := rr.Kaoto.Spec.Image
	if image == "" {
		image = defaults.KaotoStandaloneImage
	}

	labels := Labels(rr.Kaoto)

	envs := make([]*corev1ac.EnvVarApplyConfiguration, 0)
	envs = append(envs, apply.WithEnvFromField("NAMESPACE", "metadata.namespace"))

	// it appears that even with the standalone mode, there are some CORS issues that prevent some endpoint
	// to be invoked, the QUARKUS_HTTP_CORS_ORIGINS env must be set.
	//
	// TODO: investigate why this is needed
	// TODO: add support for vanilla ingress
	// TODO: ideally the operator could wait to deploy the pod till the ingress/route is deployed and the host
	//       that should be used for QUARKUS_HTTP_CORS_ORIGINS is known

	if rr.Kaoto.Spec.Ingress != nil {
		switch {
		case rr.Kaoto.Spec.Ingress.Host != "":
			// use the provided host if configured so no further lookup would be needed
			envs = append(envs, apply.WithEnv("QUARKUS_HTTP_CORS_ORIGINS", rr.Kaoto.Spec.Ingress.Host))

		case rr.ClusterType == ClusterTypeOpenShift:
			// in case of OpenShift se can leverage the Route.Spec.Host to retrieve the
			// right value for the QUARKUS_HTTP_CORS_ORIGINS
			//

			var in routev1.Route

			err := rr.Get(ctx, rr.NamespacedName, &in)
			if err != nil && !k8serrors.IsNotFound(err) {
				return nil, err
			}

			if in.Spec.Host != "" {
				host := in.Spec.Host
				if !strings.HasPrefix(host, "https://") {
					host = "https://" + host
				}

				envs = append(envs, apply.WithEnv("QUARKUS_HTTP_CORS_ORIGINS", host))
			}
		}
	}

	resource := appsv1ac.Deployment(rr.Kaoto.Name, rr.Kaoto.Namespace).
		WithOwnerReferences(apply.WithOwnerReference(rr.Kaoto)).
		WithSpec(appsv1ac.DeploymentSpec().
			WithReplicas(1).
			WithSelector(metav1ac.LabelSelector().WithMatchLabels(labels)).
			WithTemplate(corev1ac.PodTemplateSpec().
				WithLabels(labels).
				WithSpec(corev1ac.PodSpec().
					WithServiceAccountName(rr.Kaoto.Name).
					WithContainers(corev1ac.Container().
						WithImage(image).
						WithImagePullPolicy(corev1.PullAlways).
						WithName(KaotoStandaloneName).
						WithPorts(apply.WithPort(KaotoPortType, KaotoPort)).
						WithReadinessProbe(apply.WithHTTPProbe(KaotoReadinessProbePath, KaotoPort)).
						WithLivenessProbe(apply.WithHTTPProbe(KaotoLivenessProbePath, KaotoPort)).
						WithEnv(envs...).
						WithResources(corev1ac.ResourceRequirements().WithRequests(corev1.ResourceList{
							corev1.ResourceMemory: KaotoStandaloneDefaultMemory,
							corev1.ResourceCPU:    KaotoStandaloneDefaultCPU,
						}))))))

	return resource, nil
}
