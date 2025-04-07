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

type Workspace struct {
	Id              int64  `json:"id,omitempty" gorm:"column:id;primaryKey;AUTO_INCREMENT"`
	Name            string `json:"name,omitempty" gorm:"column:name;default:'';NOT NULL"`
	Description     string `json:"description,omitempty" gorm:"column:namespace;default:'';NOT NULL"`
	ClusterId       int64  `json:"cluster_id,omitempty" gorm:"column:cluster_id;default:0;NOT NULL"`
	UserId          int64  `json:"user_id,omitempty" gorm:"column:user_id;default:0;NOT NULL"`
	CpuRate         int32  `json:"cpu_rate,omitempty" gorm:"column:cpu_rate;default:0;NOT NULL"`
	GpuRate         int32  `json:"gpu_rate,omitempty" gorm:"column:gpu_rate;default:0;NOT NULL"`
	MemoryRate      int32  `json:"memory_rate,omitempty" gorm:"column:memory_rate;default:0;NOT NULL"`
	DiskRate        int32  `json:"disk_rate,omitempty" gorm:"column:disk_rate;default:0;NOT NULL"`
	LimitCpu        int32  `json:"limit_cpu,omitempty" gorm:"column:limit_cpu;default:0;NOT NULL"`
	LimitGpu        int32  `json:"limit_gpu,omitempty" gorm:"column:limit_gpu;default:0;NOT NULL"`
	LimitMemory     int32  `json:"limit_memory,omitempty" gorm:"column:limit_memory;default:0;NOT NULL"`
	LimitDisk       int32  `json:"limit_disk,omitempty" gorm:"column:limit_disk;default:0;NOT NULL"`
	GitRepository   string `json:"git_repository,omitempty" gorm:"column:git_repository;default:'';NOT NULL"`
	ImageRepository string `json:"image_repository,omitempty" gorm:"column:image_repository;default:'';NOT NULL"`
}

// // CloudWorkspaceSpec defines the desired state of CloudWorkspace.
// type CloudWorkspaceSpec struct {
// 	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file

//		// Foo is an example field of CloudWorkspace. Edit cloudworkspace_types.go to remove/update
//		Foo string `json:"foo,omitempty"`
//	}
type CloudWorkspaceStatusValue int32

const (
	CloudWorkspaceStatusValuePending   = 0
	CloudWorkspaceStatusValueCompleted = 1
	CloudWorkspaceStatusValueFailed    = 2
)

// CloudWorkspaceStatus defines the observed state of CloudWorkspace.
type CloudWorkspaceStatus struct {
	Status CloudWorkspaceStatusValue `json:"status,omitempty"`
	Meaage string                    `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudWorkspace is the Schema for the cloudworkspaces API.
type CloudWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Workspace            `json:"spec,omitempty"`
	Status CloudWorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudWorkspaceList contains a list of CloudWorkspace.
type CloudWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudWorkspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudWorkspace{}, &CloudWorkspaceList{})
}
