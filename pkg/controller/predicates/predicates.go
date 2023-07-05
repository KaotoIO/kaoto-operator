package predicates

import (
	"reflect"

	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var Log = logf.Log.WithName("predicate")

// StatusChanged implements a generic update predicate function on status change.
type StatusChanged struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating status change.
func (in StatusChanged) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		Log.Error(nil, "Update event has no old object to update", "event", e)
		return false
	}
	if e.ObjectNew == nil {
		Log.Error(nil, "Update event has no new object to update", "event", e)
		return false
	}

	s1 := reflect.ValueOf(e.ObjectOld).Elem().FieldByName("Status")
	if !s1.IsValid() {
		Log.Error(nil, "Update event old object has no Status field", "event", e)
		return false
	}

	s2 := reflect.ValueOf(e.ObjectNew).Elem().FieldByName("Status")
	if !s2.IsValid() {
		Log.Error(nil, "Update event new object has no Status field", "event", e)
		return false
	}

	return !equality.Semantic.DeepEqual(s1.Interface(), s2.Interface())
}

type AnnotationChanged struct {
	predicate.Funcs
	Name string
}

// Update implements default UpdateEvent filter for validating annotation change.
func (in AnnotationChanged) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil {
		Log.Error(nil, "Update event has no old object to update", "event", e)
		return false
	}
	if e.ObjectOld.GetAnnotations() == nil {
		Log.Error(nil, "Update event has no old object annotations to update", "event", e)
		return false
	}

	if e.ObjectNew == nil {
		Log.Error(nil, "Update event has no new object for update", "event", e)
		return false
	}

	if e.ObjectNew.GetAnnotations() == nil {
		Log.Error(nil, "Update event has no new object annotations for update", "event", e)
		return false
	}

	oldAnnotations := e.ObjectOld.GetAnnotations()
	newAnnotations := e.ObjectNew.GetAnnotations()

	return oldAnnotations[in.Name] != newAnnotations[in.Name]
}
