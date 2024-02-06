package designer

import (
	"context"

	"github.com/kaotoIO/kaoto-operator/pkg/apply"
	"github.com/kaotoIO/kaoto-operator/pkg/controller/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kaotoIO/kaoto-operator/pkg/defaults"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
)

func NewDeployAction() Action {
	return &deployAction{}
}

type deployAction struct {
}

func (a *deployAction) Configure(_ context.Context, _ *client.Client, b *builder.Builder) (*builder.Builder, error) {
	b = b.Owns(&appsv1.Deployment{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))

	return b, nil
}

func (a *deployAction) Cleanup(context.Context, *ReconciliationRequest) error {
	return nil
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

	d := a.deployment(rr)

	_, err := rr.Client.AppsV1().Deployments(rr.Kaoto.Namespace).Apply(
		ctx,
		d,
		metav1.ApplyOptions{
			FieldManager: KaotoOperatorFieldManager,
			Force:        true,
		},
	)

	return err
}

func (a *deployAction) deployment(rr *ReconciliationRequest) *appsv1ac.DeploymentApplyConfiguration {
	image := rr.Kaoto.Spec.Image
	if image == "" {
		image = defaults.KaotoAppImage
	}

	labels := Labels(rr.Kaoto)

	envs := make([]*corev1ac.EnvVarApplyConfiguration, 0)
	envs = append(envs, apply.WithEnvFromField("NAMESPACE", "metadata.namespace"))

	return appsv1ac.Deployment(rr.Kaoto.Name, rr.Kaoto.Namespace).
		WithOwnerReferences(apply.WithOwnerReference(rr.Kaoto)).
		WithLabels(Labels(rr.Kaoto)).
		WithSpec(appsv1ac.DeploymentSpec().
			WithReplicas(1).
			WithSelector(metav1ac.LabelSelector().WithMatchLabels(labels)).
			WithTemplate(corev1ac.PodTemplateSpec().
				WithLabels(labels).
				WithSpec(corev1ac.PodSpec().
					WithSecurityContext(corev1ac.PodSecurityContext().
						WithRunAsNonRoot(true).
						WithSeccompProfile(corev1ac.SeccompProfile().WithType(corev1.SeccompProfileTypeRuntimeDefault))).
					WithContainers(corev1ac.Container().
						WithImage(image).
						WithImagePullPolicy(corev1.PullAlways).
						WithName(KaotoAppName).
						WithPorts(apply.WithPort(KaotoPortType, KaotoPort)).
						WithReadinessProbe(apply.WithHTTPProbe(KaotoReadinessProbePath, KaotoPort)).
						WithLivenessProbe(apply.WithHTTPProbe(KaotoLivenessProbePath, KaotoPort)).
						WithEnv(envs...).
						WithResources(corev1ac.ResourceRequirements().WithRequests(corev1.ResourceList{
							corev1.ResourceMemory: KaotoStandaloneDefaultMemory,
							corev1.ResourceCPU:    KaotoStandaloneDefaultCPU,
						})).
						WithSecurityContext(corev1ac.SecurityContext().
							WithAllowPrivilegeEscalation(false).
							WithRunAsNonRoot(true).
							WithSeccompProfile(corev1ac.SeccompProfile().WithType(corev1.SeccompProfileTypeRuntimeDefault)))))))
}
