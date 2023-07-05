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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *KaotoReconciler) doReconcile(ctx context.Context, rr *ReconciliationRequest) (ctrl.Result, error) {

	//
	// Frontend
	//

	{
		frontendDep := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name + "-frontend",
				Namespace: rr.Kaoto.Namespace,
			},
		}

		condition := metav1.Condition{
			Type:               "FrontendDeployment",
			Status:             metav1.ConditionTrue,
			Reason:             "Deployed",
			Message:            "Deployed",
			ObservedGeneration: rr.Kaoto.Generation,
		}

		if err := reify(
			ctx,
			rr,
			&frontendDep,
			func(resource *appsv1.Deployment) *appsv1.Deployment {
				return resource.DeepCopy()
			},
			func(resource *appsv1.Deployment) *appsv1.Deployment {
				return GetFrontEndDeployment(rr.Kaoto, resource)
			},
		); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "Failure"
			condition.Message = err.Error()
		}

		meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, condition)
	}

	{
		frontendSvc := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name + "-frontend",
				Namespace: rr.Kaoto.Namespace,
			},
		}

		condition := metav1.Condition{
			Type:               "FrontendService",
			Status:             metav1.ConditionTrue,
			Reason:             "Deployed",
			Message:            "Deployed",
			ObservedGeneration: rr.Kaoto.Generation,
		}

		if err := reify(
			ctx,
			rr,
			&frontendSvc,
			func(resource *v1.Service) *v1.Service {
				return resource.DeepCopy()
			},
			func(resource *v1.Service) *v1.Service {
				return NewService(resource, rr.Kaoto.Name, rr.Kaoto.Spec.Frontend.Name, rr.Kaoto.Namespace, 80, rr.Kaoto.Spec.Frontend.Port)
			},
		); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "Failure"
			condition.Message = err.Error()
		}

		meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, condition)
	}

	//
	// Backend
	//

	{
		backendDep := appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name + "-backend",
				Namespace: rr.Kaoto.Namespace,
			},
		}

		condition := metav1.Condition{
			Type:               "BackedDeployment",
			Status:             metav1.ConditionTrue,
			Reason:             "Deployed",
			Message:            "Deployed",
			ObservedGeneration: rr.Kaoto.Generation,
		}

		if err := reify(
			ctx,
			rr,
			&backendDep,
			func(resource *appsv1.Deployment) *appsv1.Deployment {
				return resource.DeepCopy()
			},
			func(resource *appsv1.Deployment) *appsv1.Deployment {
				return GetBackendDeployment(rr.Kaoto, nil, resource)
			},
		); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "Failure"
			condition.Message = err.Error()
		}

		meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, condition)
	}

	{
		backendSvc := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rr.Kaoto.Name + "-backend",
				Namespace: rr.Kaoto.Namespace,
			},
		}

		condition := metav1.Condition{
			Type:               "BckendService",
			Status:             metav1.ConditionTrue,
			Reason:             "Deployed",
			Message:            "Deployed",
			ObservedGeneration: rr.Kaoto.Generation,
		}

		if err := reify(
			ctx,
			rr,
			&backendSvc,
			func(resource *v1.Service) *v1.Service {
				return resource.DeepCopy()
			},
			func(resource *v1.Service) *v1.Service {
				return NewService(resource, rr.Kaoto.Name, rr.Kaoto.Spec.Backend.Name, rr.Kaoto.Namespace, rr.Kaoto.Spec.Backend.Port, rr.Kaoto.Spec.Frontend.Port)
			},
		); err != nil {
			condition.Status = metav1.ConditionFalse
			condition.Reason = "Failure"
			condition.Message = err.Error()
		}

		meta.SetStatusCondition(&rr.Kaoto.Status.Conditions, condition)
	}

	return ctrl.Result{}, nil
}

func reify[T client.Object](
	ctx context.Context,
	rr *ReconciliationRequest,
	obj T,
	copier func(T) T,
	action func(T) T) error {

	if err := rr.GetDependant(obj); err != nil {
		return fmt.Errorf(
			"unable to get resource (namespace: %s, name: %s): %s",
			obj.GetNamespace(),
			obj.GetName(),
			err.Error())
	}

	source := copier(obj)
	target := action(obj)

	if _, err := rr.PatchDependant(source, target); err != nil {
		return fmt.Errorf(
			"unable to patch resource (namespace: %s, name: %s): %s",
			obj.GetNamespace(),
			obj.GetName(),
			err.Error())
	}

	return nil
}
