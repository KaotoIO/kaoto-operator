package controllers

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewService(appName, serviceMame, namespace string, port, targetPort int32) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    map[string]string{"app": appName},
			Name:      serviceMame + "-svc",
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       serviceMame + "port",
					Protocol:   "TCP",
					Port:       port,
					TargetPort: intstr.FromInt(int(targetPort)),
				},
			},
			Selector:                 map[string]string{"app": serviceMame},
			SessionAffinity:          "None",
			PublishNotReadyAddresses: true,
		},
	}
}
