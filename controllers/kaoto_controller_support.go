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
	"sort"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/api/v1alpha1"
	"github.com/kaotoIO/kaoto-operator/pkg/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// KaotoReconciler reconciles a Kaoto object
type KaotoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *KaotoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr)

	c = c.For(&kaotoiov1alpha1.Kaoto{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
			predicate.LabelChangedPredicate{},
		)))

	c = c.Owns(&appsv1.Deployment{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
			predicate.LabelChangedPredicate{},
		)))

	c = c.Owns(&v1.Service{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
			predicate.LabelChangedPredicate{},
		)))

	/*
		dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
		if err != nil {
			return err
		}

		ok, err := openshift.IsOpenShift(dc)
		if err != nil {
			return err
		}
		if ok {
			c.Owns(&routev1.Route{})
		}
	*/

	return c.Complete(r)
}

//+kubebuilder:rbac:groups=kaoto.io,namespace=kaoto-operator,resources=kaotoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kaoto.io,namespace=kaoto-operator,resources=kaotoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kaoto.io,namespace=kaoto-operator,resources=kaotoes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,namespace=kaoto-operator,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",namespace=kaoto-operator,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="route.openshift.io",namespace=kaoto-operator,resources=routes,verbs=get;list;watch;create;update;patch;delete

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
	l := log.FromContext(ctx)
	l.Info("Reconciling", "resource", req.NamespacedName.String())

	kaotoRef := &kaotoiov1alpha1.Kaoto{}
	err := r.Get(ctx, req.NamespacedName, kaotoRef)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// no CR found
			return ctrl.Result{}, nil
		}
	}

	rr := ReconciliationRequest{
		C:      ctx,
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		// safety copy
		Kaoto: kaotoRef.DeepCopy(),
	}

	if rr.Kaoto.ObjectMeta.DeletionTimestamp.IsZero() {

		//
		// Add finalizer
		//

		if controllerutil.AddFinalizer(rr.Kaoto, defaults.KaotoFinalizerName) {
			if err := r.Update(ctx, rr.Kaoto); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure adding finalizer to connector cluster %s", req.NamespacedName)
			}
		}
	} else {

		//
		// Handle deletion
		//

		if controllerutil.RemoveFinalizer(rr.Kaoto, defaults.KaotoFinalizerName) {
			if err := r.Update(ctx, rr.Kaoto); err != nil {
				if k8serrors.IsConflict(err) {
					return ctrl.Result{}, err
				}

				return ctrl.Result{}, errors.Wrapf(err, "failure removing finalizer from connector cluster %s", req.NamespacedName)
			}
		}

		return ctrl.Result{}, nil
	}

	rr.Kaoto.Status.ObservedGeneration = rr.Kaoto.Generation
	rr.Kaoto.Status.Phase = "Running"

	//
	// Reconcilie
	//

	result, err := r.doReconcile(ctx, &rr)

	// TODO: must be refactored
	if err != nil || result.Requeue == true || !result.IsZero() {
		return result, err
	}

	//
	// update status
	//

	sort.SliceStable(rr.Kaoto.Status.Conditions, func(i, j int) bool {
		return rr.Kaoto.Status.Conditions[i].Type < rr.Kaoto.Status.Conditions[j].Type
	})

	if err := r.Status().Update(ctx, rr.Kaoto); err != nil {
		if k8serrors.IsConflict(err) {
			l.Info(err.Error())
			return ctrl.Result{Requeue: true}, nil
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
