/*
Copyright 2023.

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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CronJobTemplateSpec struct {
	// Standard object's metadata of the jobs created from this template.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the job.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec batchv1.CronJobSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// CronSetSpec defines the desired state of CronSet
type CronSetSpec struct {
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,1,opt,name=selector"`

	CronJobTemplate CronJobTemplateSpec `json:"template,omitempty" protobuf:"bytes,2,opt,name=template"`
}

// CronSetStatus defines the observed state of CronSet
type CronSetStatus struct {
	CurrentNumberScheduled int32 `json:"currentNumberScheduled" protobuf:"varint,1,opt,name=currentNumberScheduled"`

	NumberMisscheduled int32 `json:"numberMisscheduled" protobuf:"varint,2,opt,name=numberMisscheduled"`

	DesiredNumberScheduled int32 `json:"desiredNumberScheduled" protobuf:"varint,3,opt,name=desiredNumberScheduled"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CronSet is the Schema for the cronsets API
type CronSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronSetSpec   `json:"spec,omitempty"`
	Status CronSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CronSetList contains a list of CronSet
type CronSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronSet{}, &CronSetList{})
}
