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

type AppReleaseSatus int32

const (
	AppReleaseSatus_UNSPECIFIED AppReleaseSatus = 0
	AppReleaseSatus_PENDING     AppReleaseSatus = 1
	AppReleaseSatus_RUNNING     AppReleaseSatus = 2
	AppReleaseSatus_FAILED      AppReleaseSatus = 3
)

type AppReleaseResourceStatus int32

const (
	AppReleaseResourceStatus_AppReleaseResourceStatus_UNSPECIFIED AppReleaseResourceStatus = 0
	AppReleaseResourceStatus_SUCCESSFUL                           AppReleaseResourceStatus = 1
	AppReleaseResourceStatus_UNHEALTHY                            AppReleaseResourceStatus = 2
)

type AppReleaseResource struct {
	Id        string                   `json:"id,omitempty"`
	ReleaseId int64                    `json:"release_id,omitempty"`
	Name      string                   `json:"name,omitempty"`
	Namespace string                   `json:"namespace,omitempty"`
	Kind      string                   `json:"kind,omitempty"`
	Lables    string                   `json:"lables,omitempty"`
	Manifest  string                   `json:"manifest,omitempty"`
	StartedAt string                   `json:"started_at,omitempty"`
	Status    AppReleaseResourceStatus `json:"status,omitempty"`
	Events    string                   `json:"events,omitempty"`
}

type AppRelease struct {
	Id          int64                 `json:"id,omitempty"`
	AppVersion  string                `json:"app_version,omitempty"`
	AppName     string                `json:"app_name,omitempty"`
	ReleaseName string                `json:"release_name,omitempty"`
	Namespace   string                `json:"namespace,omitempty"`
	Config      string                `json:"config,omitempty"`
	ConfigFile  string                `json:"config_file,omitempty"`
	Status      AppReleaseSatus       `json:"status,omitempty"`
	Notes       string                `json:"notes,omitempty"`
	Logs        string                `json:"logs,omitempty"`
	Dryrun      bool                  `json:"dryrun,omitempty"`
	Atomic      bool                  `json:"atomic,omitempty"`
	Wait        bool                  `json:"wait,omitempty"`
	Replicas    int32                 `json:"replicas,omitempty"`
	Cpu         int32                 `json:"cpu,omitempty"`
	LimitCpu    int32                 `json:"limit_cpu,omitempty"`
	Memory      int32                 `json:"memory,omitempty"`
	LimitMemory int32                 `json:"limit_memory,omitempty"`
	Gpu         int32                 `json:"gpu,omitempty"`
	LimitGpu    int32                 `json:"limit_gpu,omitempty"`
	Storage     int32                 `json:"storage,omitempty"`
	Resources   []*AppReleaseResource `json:"resources,omitempty"`
	Chart       string                `json:"chart,omitempty"`
	RepoName    string                `json:"repo_name,omitempty"`
	AppId       int64                 `json:"app_id,omitempty"`
	VersionId   int64                 `json:"version_id,omitempty"`
	ClusterId   int64                 `json:"cluster_id,omitempty"`
	ProjectId   int64                 `json:"project_id,omitempty"`
	UserId      int64                 `json:"user_id,omitempty"`
	WorkspaceId int64                 `json:"workspace_id,omitempty"`
}

// CloudAppReleaseStatus defines the observed state of CloudAppRelease.
type CloudAppReleaseStatus struct {
	Status AppReleaseSatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudAppRelease is the Schema for the cloudappreleases API.
type CloudAppRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppRelease            `json:"spec,omitempty"`
	Status CloudAppReleaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudAppReleaseList contains a list of CloudAppRelease.
type CloudAppReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudAppRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudAppRelease{}, &CloudAppReleaseList{})
}
