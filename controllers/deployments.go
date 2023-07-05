package controllers

import (
	"github.com/kaotoIO/kaoto-operator/api/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetFrontEndDeployment(kaoto *v1alpha1.Kaoto, deployment *appsv1.Deployment) *appsv1.Deployment {
	image := kaoto.Spec.Frontend.Image
	vars := []corev1.EnvVar{}

	getDeployment(kaoto.Name, kaoto.Spec.Frontend.Name, kaoto.Namespace, kaoto.Spec.Frontend.Name, image, kaoto.Spec.Frontend.Port, "default", vars, deployment)

	return deployment
}

func GetBackendDeployment(kaoto *v1alpha1.Kaoto, kaotoRoute *routev1.Route, deployment *appsv1.Deployment) *appsv1.Deployment {
	image := kaoto.Spec.Backend.Image
	vars := make([]corev1.EnvVar, 0)

	vars = append(vars, corev1.EnvVar{
		Name:  "NAMESPACE",
		Value: kaoto.Namespace,
	})

	if kaotoRoute != nil {
		vars = append(vars, corev1.EnvVar{
			Name:  "QUARKUS_HTTP_CORS_ORIGINS",
			Value: "https://" + kaotoRoute.Spec.Host,
		})
	}

	getDeployment(kaoto.Name, kaoto.Spec.Backend.Name, kaoto.Namespace, kaoto.Spec.Backend.Name, image, kaoto.Spec.Backend.Port, "kaoto-operator-integrator-sa", vars, deployment)

	return deployment

}

func getDeployment(kaotoName, name, namespace, imageName, image string, port int32, saName string, vars []corev1.EnvVar, deployment *appsv1.Deployment) {

	labels := labelsForKaoto(name, kaotoName)
	replicas := int32(1)
	deployment.Spec = appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: saName,
				Containers: []corev1.Container{{
					Image: image,
					Name:  imageName,
					Env:   vars,
					Ports: []corev1.ContainerPort{{
						ContainerPort: port,
						Name:          "port",
					}},
					ImagePullPolicy: "Always",
				}},
			},
		},
	}
}
func labelsForKaoto(app, name string) map[string]string {
	return map[string]string{"app": app, "kaoto_cr": name}
}
