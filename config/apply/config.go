package apply

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WithOwnerReference(object client.Object) *metav1ac.OwnerReferenceApplyConfiguration {
	return metav1ac.OwnerReference().
		WithAPIVersion(object.GetObjectKind().GroupVersionKind().GroupVersion().String()).
		WithKind(object.GetObjectKind().GroupVersionKind().Kind).
		WithName(object.GetName()).
		WithUID(object.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true)
}

func WithHTTPProbe(path string, port int32) *corev1ac.ProbeApplyConfiguration {
	return corev1ac.Probe().
		WithInitialDelaySeconds(5).
		WithPeriodSeconds(1).
		WithFailureThreshold(10).
		WithSuccessThreshold(1).
		WithTimeoutSeconds(10).
		WithHTTPGet(corev1ac.HTTPGetAction().
			WithPath(path).
			WithPort(intstr.IntOrString{IntVal: port}).
			WithScheme(corev1.URISchemeHTTP))
}

func WithPort(name string, port int32) *corev1ac.ContainerPortApplyConfiguration {
	return corev1ac.ContainerPort().
		WithName(name).
		WithContainerPort(port).
		WithProtocol(corev1.ProtocolTCP)
}

func WithEnv(name string, value string) *corev1ac.EnvVarApplyConfiguration {
	return corev1ac.EnvVar().
		WithName(name).
		WithValue(value)
}

func WithEnvFromField(name string, value string) *corev1ac.EnvVarApplyConfiguration {
	return corev1ac.EnvVar().
		WithName(name).
		WithValueFrom(corev1ac.EnvVarSource().WithFieldRef(corev1ac.ObjectFieldSelector().WithFieldPath(value)))
}
