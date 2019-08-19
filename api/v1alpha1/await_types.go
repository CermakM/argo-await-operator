/*
Copyright 2019 Marek Cermak <macermak@redhat.com>.

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

// +kubebuilder:object:root=true

// Await is the Schema for the awaits API
type Await struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwaitSpec   `json:"spec,omitempty"`
	Status AwaitStatus `json:"status,omitempty"`
}

// AwaitSpec defines the desired state of Await
// +k8s:openapi-gen=true
type AwaitSpec struct {
	Workflow NamespacedWorkflow  `json:"workflow"`
	Resource *metav1.APIResource `json:"resource"`
	Filters  []string            `json:"filters,omitempty"`
}

// NamespacedWorkflow defines the workflow to be resumed
type NamespacedWorkflow struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// AwaitStatus defines the observed state of Await
// +k8s:openapi-gen=true
type AwaitStatus struct {
	StartedAt  metav1.Time `json:"startedAt,omitempty"`
	FinishedAt metav1.Time `json:"finishedAt,omitempty"`
}

// +kubebuilder:object:root=true

// AwaitList contains a list of Await
type AwaitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Await `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Await{}, &AwaitList{})
}
