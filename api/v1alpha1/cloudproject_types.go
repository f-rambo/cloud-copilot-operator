/*
Copyright 2025.

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

// CloudProjectSpec defines the desired state of CloudProject.
// type CloudProjectSpec struct {
// 	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file

// 	// Foo is an example field of CloudProject. Edit cloudproject_types.go to remove/update
// 	Foo string `json:"foo,omitempty"`
// }

type Project struct {
	Id          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Description string `json:"description,omitempty"`
	ClusterId   int64  `json:"cluster_id,omitempty"`
	UserId      int64  `json:"user_id,omitempty"`
	WorkspaceId int64  `json:"workspace_id,omitempty"`
	LimitCpu    int32  `json:"limit_cpu,omitempty"`
	LimitGpu    int32  `json:"limit_gpu,omitempty"`
	LimitMemory int32  `json:"limit_memory,omitempty"`
	LimitDisk   int32  `json:"limit_disk,omitempty"`
}

type ProjectStatus int32

const (
	ProjectStatusPending ProjectStatus = 0
	ProjectStatusRunning ProjectStatus = 1
	ProjectStatusFailed  ProjectStatus = 2
)

// CloudProjectStatus defines the observed state of CloudProject.
type CloudProjectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  ProjectStatus `json:"status,omitempty"`
	Message string        `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudProject is the Schema for the cloudprojects API.
type CloudProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Project            `json:"spec,omitempty"`
	Status CloudProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudProjectList contains a list of CloudProject.
type CloudProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudProject{}, &CloudProjectList{})
}
