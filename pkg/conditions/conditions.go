package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ConditionType is a valid value for Condition.Type.
type ConditionType string

// Getter interface defines methods that an object should implement in order to
// use the conditions package for getting conditions.
type Getter interface {
	metav1.Object
	runtime.Object

	// GetConditions returns the list of conditions for an object.
	GetConditions() Conditions
}

type Conditions []metav1.Condition

// Get returns the condition with the given type, if the condition does not exist,
// it returns nil.
func Get(from Getter, t ConditionType) *metav1.Condition {
	conditions := from.GetConditions()
	if conditions == nil {
		return nil
	}

	for _, condition := range conditions {
		if condition.Type == string(t) {
			return &condition
		}
	}
	return nil
}
