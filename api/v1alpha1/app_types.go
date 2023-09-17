/*
Copyright 2023 f-rambo.

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

// AppSpec defines the desired state of App
type AppSpec struct {
	AppChart AppChart `json:"appChart,omitempty"`
	Service  Service  `json:"service,omitempty"`
}

type AppChart struct {
	Enable        bool   `json:"enable"`
	RepoName      string `json:"repoName,omitempty"`
	RepoURL       string `json:"repoURL,omitempty"`
	ChartName     string `json:"chartName,omitempty"`
	Version       string `json:"version,omitempty"`
	ConfigMapName string `json:"configMapName,omitempty"`
	SecretName    string `json:"secretName,omitempty"`
}

type Service struct {
	Enable        bool    `json:"enable"`
	EnableIngress bool    `json:"enable_ingress,omitempty"`
	EnableService bool    `json:"enable_service"`
	Replicas      int32   `json:"replicas"`
	Registry      string  `json:"registry,omitempty"`
	Image         string  `json:"image,omitempty"`
	Version       string  `json:"version,omitempty"`
	CPU           string  `json:"cpu,omitempty"`
	GPU           string  `json:"gpa,omitempty"`
	Memory        string  `json:"memory,omitempty"`
	ConfigMapName string  `json:"configMapName,omitempty"`
	Ports         []int32 `json:"ports,omitempty"`
}

// AppStatus defines the observed state of App
type AppStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="repoName",type=string,JSONPath=`.spec.repoName`
//+kubebuilder:printcolumn:name="repoURL",type=string,JSONPath=`.spec.repoURL`
//+kubebuilder:printcolumn:name="chartName",type=string,JSONPath=`.spec.chartName`
//+kubebuilder:printcolumn:name="version",type=string,JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="configMapName",type=string,JSONPath=`.spec.configMapName`
//+kubebuilder:printcolumn:name="secretName",type=string,JSONPath=`.spec.secretName`

// App is the Schema for the apps API
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec   `json:"spec,omitempty"`
	Status AppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AppList contains a list of App
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}
