package resources

import (
	"context"

	"github.com/kaotoIO/kaoto-operator/pkg/patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func SetAnnotation(target ctrl.Object, key string, val string) {
	if target.GetAnnotations() == nil {
		target.SetAnnotations(make(map[string]string))
	}

	if key != "" && val != "" {
		target.GetAnnotations()[key] = val
	}
}

func SetAnnotations(target ctrl.Object, values map[string]string) {
	if target.GetAnnotations() == nil {
		target.SetAnnotations(make(map[string]string))
	}

	for k, v := range values {
		target.GetAnnotations()[k] = v
	}
}

func SetLabel(target *metav1.ObjectMeta, key string, val string) {
	if target.GetLabels() == nil {
		target.SetLabels(make(map[string]string))
	}

	if key != "" && val != "" {
		target.GetLabels()[key] = val
	}
}

func SetLabels(target ctrl.Object, values map[string]string) {
	if target.GetLabels() == nil {
		target.SetLabels(make(map[string]string))
	}

	for k, v := range values {
		target.GetLabels()[k] = v
	}
}

func Apply(
	ctx context.Context,
	c ctrl.Client,
	target ctrl.Object,
) error {
	data, err := patch.ApplyPatch(target)
	if err != nil {
		return err
	}

	return c.Patch(ctx, data, ctrl.Apply, ctrl.ForceOwnership, ctrl.FieldOwner("kaoto-operator"))
}

func AsNamespacedName(obj ctrl.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func Get(ctx context.Context, c ctrl.Reader, target ctrl.Object, opts ...ctrl.GetOption) error {
	return c.Get(ctx, AsNamespacedName(target), target, opts...)
}

func AsObjectMeta(resource types.NamespacedName) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      resource.Name,
		Namespace: resource.Namespace,
	}
}
