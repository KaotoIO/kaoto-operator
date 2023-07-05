package controllers

import (
	"context"

	kaotoiov1alpha1 "github.com/kaotoIO/kaoto-operator/api/v1alpha1"
	"github.com/kaotoIO/kaoto-operator/pkg/resources"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ReconciliationRequest struct {
	client.Client
	types.NamespacedName

	C context.Context

	Kaoto *kaotoiov1alpha1.Kaoto
}

func (rc *ReconciliationRequest) PatchDependant(source client.Object, target client.Object) (bool, error) {

	if err := controllerutil.SetControllerReference(rc.Kaoto, target, rc.Scheme()); err != nil {
		return false, errors.Wrapf(err, "unable to set controller reference to: %s", target.GetObjectKind().GroupVersionKind().String())
	}

	if target.GetAnnotations() == nil {
		target.SetAnnotations(make(map[string]string))
	}

	return resources.Apply(rc.C, rc.Client, source, target)
}

func (rc *ReconciliationRequest) GetDependant(obj client.Object, opts ...client.GetOption) error {
	nn := rc.NamespacedName
	if obj.GetNamespace() != "" {
		nn.Namespace = obj.GetNamespace()
	}
	if obj.GetName() != "" {
		nn.Name = obj.GetName()
	}

	err := rc.Client.Get(rc.C, nn, obj, opts...)
	if k8serrors.IsNotFound(err) {
		obj.SetName(nn.Name)
		obj.SetNamespace(nn.Namespace)

		return nil
	}

	return err
}

func (rc *ReconciliationRequest) DeleteDependant(obj client.Object, opts ...client.DeleteOption) error {
	return rc.Client.Delete(rc.C, obj, opts...)
}



func (rc *ReconciliationRequest) Reify[T client.Object](
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