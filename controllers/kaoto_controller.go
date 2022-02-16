/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/api/v1alpha1"
)

// KaotoReconciler reconciles a Kaoto object
type KaotoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kaoto.io,resources=kaotoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kaoto.io,resources=kaotoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kaoto.io,resources=kaotoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kaoto object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *KaotoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	kaoto := &kaotoiov1alpha1.Kaoto{}
	err := r.Get(ctx, req.NamespacedName, kaoto)
	if err != nil {
		log.Error(err, "problemis")
		if errors.IsNotFound(err) {
			// no CR found
			return ctrl.Result{}, nil
		}
	}

	backendName := kaoto.Name + "-backend"
	backendDep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: backendName, Namespace: kaoto.Namespace}, backendDep)
	if err != nil && errors.IsNotFound(err) {
		backend := kaoto.Spec.Backend
		backendDep = r.getDeployment(backendName, kaoto.Namespace, "kaoto-backend", backend.Image, backend.Port)
		log.Info("Creating a new Deployment")
		err = r.Create(ctx, backendDep)

		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", backendDep.Namespace, "Deployment.Name", backendDep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KaotoReconciler) getDeployment(name, namespace, imageName, image string, port int32) *appsv1.Deployment {
	ls := labelsForKaotoBackend(name)
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: image,
						Name:  imageName,
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
							Name:          "port",
						}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	return dep
}

func labelsForKaotoBackend(name string) map[string]string {
	return map[string]string{"app": "kaoto-backend", "kaoto_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KaotoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kaotoiov1alpha1.Kaoto{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
