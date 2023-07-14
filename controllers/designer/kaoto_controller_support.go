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
	"fmt"

	"github.com/kaotoIO/kaoto-operator/pkg/resources"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func reify[T ctrl.Object](
	ctx context.Context,
	rr *ReconciliationRequest,
	obj T,
	action func(T) (T, error)) error {

	gvk, err := apiutil.GVKForObject(obj, rr.Scheme())
	if err != nil {
		return err
	}

	obj.GetObjectKind().SetGroupVersionKind(gvk)

	target, err := action(obj)
	if err != nil {
		return err
	}

	resources.SetLabels(obj, Labels(obj))

	if err := resources.Apply(ctx, rr.Client, target); err != nil {
		return fmt.Errorf(
			"unable to patch resource (namespace: %s, name: %s): %s",
			obj.GetNamespace(),
			obj.GetName(),
			err.Error())
	}

	return nil
}

func Labels(ref ctrl.Object) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "kaoto",
		"app.kubernetes.io/instance":   ref.GetName(),
		"app.kubernetes.io/component":  "designer",
		"app.kubernetes.io/part-of":    "kaoto",
		"app.kubernetes.io/managed-by": "kaoto-operator",
	}
}

func LabelsForSelector(ref ctrl.Object) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     "kaoto",
		"app.kubernetes.io/instance": ref.GetName(),
	}
}
