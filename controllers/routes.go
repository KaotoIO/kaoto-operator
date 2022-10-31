package controllers

import (
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRoute(appName, routeName string, service v1.Service) *routev1.Route {
	routeTLSConfig := &routev1.TLSConfig{
		Termination:                   routev1.TLSTerminationEdge,
		InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
	}

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    map[string]string{"app": appName},
			Name:      routeName,
			Namespace: service.Namespace,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: service.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: service.Spec.Ports[0].TargetPort,
			},
			TLS: routeTLSConfig,
		},
	}
}
