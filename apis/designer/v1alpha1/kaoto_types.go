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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// KaotoSpec defines the desired state of Kaoto.
type KaotoSpec struct {
	// +optional
	Image string `json:"image,omitempty"`

	// +optional
	Ingress *IngressSpec `json:"ingress,omitempty"`
}

type IngressSpec struct {
	// +optional
	Host string `json:"host,omitempty"`

	// +optional
	Path string `json:"path,omitempty"`
}

// KaotoStatus defines the observed state of Kaoto.
type KaotoStatus struct {
	Phase              string             `json:"phase"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Endpoint           string             `json:"endpoint,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="The phase"
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`,description="The endpoint"
// +kubebuilder:resource:path=kaotos,scope=Namespaced,shortName=kd,categories=integration;camel

// Kaoto is the Schema for the kaotos API.
type Kaoto struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KaotoSpec   `json:"spec,omitempty"`
	Status KaotoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KaotoList contains a list of Kaoto.
type KaotoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kaoto `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kaoto{}, &KaotoList{})
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
