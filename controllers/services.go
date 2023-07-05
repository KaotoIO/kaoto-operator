package controllers

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewService(service *v1.Service, appName string, serviceMame string, namespace string, port int32, targetPort int32) *v1.Service {
	service.Spec = v1.ServiceSpec{
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
	}

	return service
}
