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

package designer

import (
	"context"
	"sort"

	"github.com/kaotoIO/kaoto-operator/config/client"

	rbacv1 "k8s.io/api/rbac/v1"

	"go.uber.org/multierr"

	"github.com/kaotoIO/kaoto-operator/pkg/controller/predicates"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/apis/designer/v1alpha1"
	"github.com/kaotoIO/kaoto-operator/pkg/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func NewKaotoReconciler(manager ctrl.Manager) (*KaotoReconciler, error) {
	c, err := client.NewClient(manager.GetConfig(), manager.GetScheme(), manager.GetClient())
	if err != nil {
		return nil, err
	}

	rec := KaotoReconciler{}
	rec.Client = c
	rec.Scheme = manager.GetScheme()
	rec.ClusterType = ClusterTypeVanilla

	isOpenshift, err := c.IsOpenShift()
	if err != nil {
		return nil, err
	}
	if isOpenshift {
		rec.ClusterType = ClusterTypeOpenShift
	}

	rec.actions = make([]Action, 0)
	rec.actions = append(rec.actions, &rbacAction{})
	rec.actions = append(rec.actions, &serviceAction{})

	if isOpenshift {
		rec.actions = append(rec.actions, &routeAction{})
	} else {
		rec.actions = append(rec.actions, &ingressAction{})
	}

	rec.actions = append(rec.actions, &deployAction{})

	return &rec, nil
}

type KaotoReconciler struct {
	*client.Client

	Scheme      *runtime.Scheme
	ClusterType ClusterType
	actions     []Action
}

// +kubebuilder:rbac:groups=designer.kaoto.io,resources=kaotoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=designer.kaoto.io,resources=kaotoes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=designer.kaoto.io,resources=kaotoes/finalizers,verbs=update
// +kubebuilder:rbac:groups=camel.apache.org,resources=kameletbindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=camel.apache.org,resources=kamelets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=camel.apache.org,resources=integrations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="route.openshift.io",resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="networking.k8s.io",resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// SetupWithManager sets up the controller with the Manager.
func (r *KaotoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr)

	c = c.For(&kaotoiov1alpha1.Kaoto{}, builder.WithPredicates(
		predicate.Or(
			predicate.GenerationChangedPredicate{},
		)))
	c = c.Owns(&appsv1.Deployment{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))
	c = c.Owns(&corev1.Service{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))
	c = c.Owns(&corev1.ServiceAccount{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))
	c = c.Owns(&rbacv1.ClusterRoleBinding{}, builder.WithPredicates(
		predicate.Or(
			predicate.ResourceVersionChangedPredicate{},
		)))

	switch r.ClusterType {
	case ClusterTypeVanilla:
		c.Owns(&netv1.Ingress{}, builder.WithPredicates(
			predicate.Or(
				predicate.ResourceVersionChangedPredicate{},
				predicates.StatusChanged{},
			)))
	case ClusterTypeOpenShift:
		c.Owns(&routev1.Route{}, builder.WithPredicates(
			predicate.Or(
				predicate.ResourceVersionChangedPredicate{},
				predicates.StatusChanged{},
			)))

	}

	return c.Complete(r)
}

func (r *KaotoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling", "resource", req.NamespacedName.String())

	rr := ReconciliationRequest{
		Client: r.Client,
		NamespacedName: types.NamespacedName{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		ClusterType: r.ClusterType,
		// safety copy
		Kaoto: &kaotoiov1alpha1.Kaoto{},
	}

	err := r.Get(ctx, req.NamespacedName, rr.Kaoto)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// no CR found
			return ctrl.Result{}, nil
		}
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

	//
	// Reconcile
	//

	reconcileCondition := metav1.Condition{
		Type:               "Reconcile",
		Status:             metav1.ConditionTrue,
		Reason:             "Reconciled",
		Message:            "Reconciled",
		ObservedGeneration: rr.Kaoto.Generation,
	}

	var allErrors error

	for i := range r.actions {
		if err := r.actions[i].Apply(ctx, &rr); err != nil {
			allErrors = multierr.Append(allErrors, err)
		}
	}

	if allErrors != nil {
		reconcileCondition.Status = metav1.ConditionFalse
		reconcileCondition.Reason = "Failure"
		reconcileCondition.Message = "Failure"

		rr.Kaoto.Status.Phase = "Error"
	} else {
		rr.Kaoto.Status.ObservedGeneration = rr.Kaoto.Generation
		rr.Kaoto.Status.Phase = "Ready"
	}

	meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, reconcileCondition)

	sort.SliceStable(rr.Kaoto.Status.Conditions, func(i, j int) bool {
		return rr.Kaoto.Status.Conditions[i].Type < rr.Kaoto.Status.Conditions[j].Type
	})

	//
	// Update status
	//

	err = r.Status().Update(ctx, rr.Kaoto)
	if err != nil && k8serrors.IsConflict(err) {
		l.Info(err.Error())
		return ctrl.Result{Requeue: true}, nil
	} else {
		allErrors = multierr.Append(allErrors, err)
	}

	return ctrl.Result{}, allErrors
}
