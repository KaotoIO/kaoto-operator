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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KaotoSpec defines the desired state of Kaoto
type KaotoSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Kaoto. Edit kaoto_types.go to remove/update
	Foo      string        `json:"foo,omitempty"`
	Backend  KaotoBackend  `json:"backend,omitempty"`
	Frontend KaotoFrontend `json:"frontend,omitempty"`
}
type KaotoBackend struct {
	Port  int32  `json:"port,omitempty"`
	Image string `json:"image,omitempty"`
}

type KaotoFrontend struct {
	Port  int32  `json:"port,omitempty"`
	Image string `json:"image,omitempty"`
}

// KaotoStatus defines the observed state of Kaoto
type KaotoStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kaoto is the Schema for the kaotoes API
type Kaoto struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KaotoSpec   `json:"spec,omitempty"`
	Status KaotoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KaotoList contains a list of Kaoto
type KaotoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kaoto `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kaoto{}, &KaotoList{})
}
