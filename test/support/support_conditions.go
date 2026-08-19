package support

import (
	"github.com/kaotoIO/kaoto-operator/pkg/conditions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type conditionType interface {
	~string
}

func ConditionStatus[T conditionType](conditionType T) func(any) corev1.ConditionStatus {
	return func(object any) corev1.ConditionStatus {
		switch o := object.(type) {
		case conditions.Getter:
			if c := conditions.Get(o, conditions.ConditionType(conditionType)); c != nil {
				return corev1.ConditionStatus(c.Status)
			}
		case *appsv1.Deployment:
			return deploymentConditionStatus(o, string(conditionType))
		}

		return corev1.ConditionUnknown
	}
}

func deploymentConditionStatus(d *appsv1.Deployment, conditionType string) corev1.ConditionStatus {
	if d != nil {
		for i := range d.Status.Conditions {
			if string(d.Status.Conditions[i].Type) == conditionType {
				return d.Status.Conditions[i].Status
			}
		}
	}
	return corev1.ConditionUnknown
}

func ContainerImage(index int) func(*appsv1.Deployment) string {
	return func(deployment *appsv1.Deployment) string {
		return deployment.Spec.Template.Spec.Containers[index].Image
	}
}
