/*
Copyright 2020 The KubePreset Authors.

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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceBindingSpec defines the desired state of ServiceBinding
type ServiceBindingSpec struct {

	// Name is the name of the service as projected into the application container.  Defaults to .metadata.name.
	Name string `json:"name,omitempty"`
	// Type is the type of the service as projected into the application container
	Type string `json:"type,omitempty"`
	// Provider is the provider of the service as projected into the application container
	Provider string `json:"provider,omitempty"`

	// Application resource to inject the binding info.
	// It could be any process running within a container.
	// From the spec:
	// A Service Binding resource **MUST** define a `.spec.application`
	// which is an `ObjectReference`-like declaration to a `PodSpec`-able
	// resource.  A `ServiceBinding` **MAY** define the application
	// reference by-name or by-[label selector][ls]. A name and selector
	// **MUST NOT** be defined in the same reference.
	Application *Application `json:"application"`

	// Service referencing the binding secret
	// From the spec:
	// A Service Binding resource **MUST** define a `.spec.service` which is
	// an `ObjectReference`-like declaration to a Provisioned Service-able
	// resource.
	Service *Service `json:"service"`

	// Env creates environment variables based on the Secret values
	Env []Environment `json:"env,omitempty"`
}

type Service struct {
	// API version of the referent.
	// +optional
	APIVersion string `json:"apiVersion"`

	// Kind of the referent.
	// +optional
	Kind string `json:"kind"`

	// Name of the referent.
	// Mutually exclusive with Selector.
	// +optional
	Name string `json:"name"`
}

type Application struct {
	// API version of the referent.
	// +optional
	APIVersion string `json:"apiVersion"`

	// Kind of the referent.
	// +optional
	Kind string `json:"kind"`

	// Name of the referent.
	// Mutually exclusive with Selector.
	// +optional
	Name string `json:"name,omitempty"`

	// Selector of the referents.
	// Mutually exclusive with Name.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// ServiceBindingStatus defines the observed state of ServiceBinding
type ServiceBindingStatus struct {
	// ObservedGeneration is the 'Generation' of the Service that
	// was last processed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions the latest available observations of a resource's current state.
	// +optional
	Conditions Conditions `json:"conditions,omitempty"`

	// Binding exposes the projected secret for this ServiceBinding
	Binding *corev1.LocalObjectReference `json:"binding,omitempty"`
}

// Environment represents a key to Secret data keys and name of the environment variable
type Environment struct {
	// Name of the environment variable
	Name string `json:"name"`

	// Secret data key
	Key string `json:"key"`
}

// ConditionReady specifies that the resource is ready.
// For long-running resources.
const ConditionReady ConditionType = "Ready"

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type Conditions []Condition

type ConditionStatus string
type ConditionType string

type Condition struct {
	// Type of condition.
	// +required
	Type ConditionType `json:"type" description:"type of status condition"`

	// Status of the condition, one of True, False, Unknown.
	// +required
	Status ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`

	// LastTransitionTime is the last time the condition transitioned from one status to another.
	// We use VolatileTime in place of metav1.Time to exclude this from creating equality.Semantic
	// differences (all other things held constant).
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" description:"last time the condition transit from one status to another"`
	//LastTransitionTime VolatileTime `json:"lastTransitionTime,omitempty" description:"last time the condition transit from one status to another"`

	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" description:"one-word CamelCase reason for the condition's last transition"`

	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" description:"human-readable message indicating details about last transition"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServiceBinding is the Schema for the servicebindings API
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec,omitempty"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceBindingList contains a list of ServiceBinding
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBinding{}, &ServiceBindingList{})
}
