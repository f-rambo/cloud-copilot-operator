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
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type AppName string

func (a AppName) String() string {
	return string(a)
}

type AppCrd string

func (a AppCrd) String() string {
	return string(a)
}

const (
	DaemonSet   string = "DaemonSet"
	Deployment  string = "Deployment"
	StatefulSet string = "StatefulSet"
	ReplicaSet  string = "ReplicaSet"
	CronJob     string = "CronJob"
	Job         string = "Job"
	Pod         string = "Pod"
)

var (
	// app
	Cilium  AppName = "cilium"
	Metallb AppName = "metallb"

	RookCeph AppName = "rook-ceph"

	CertManager         AppName = "cert-manager"
	ArgoWorkflows       AppName = "argo-workflows"
	KubePrometheusStack AppName = "kube-prometheus-stack"
	EckOperator         AppName = "eck-operator"

	ClusterAutoscaler AppName = "cluster-autoscaler"
	Kafka             AppName = "kafka"

	// app crd
	GatewayCrd    AppCrd = "gateway-crd"
	ClusterIssuer AppCrd = "cluster-issuer"
	IpAddressPool AppCrd = "ip-address-pool"
	CephCluster   AppCrd = "ceph-cluster"
	StorageClass  AppCrd = "storage-class"

	// namespace
	appNsMap = map[ClusterNamespace][]AppName{
		ClusterNamespace_networking: {Metallb, Cilium},
		ClusterNamespace_storage:    {RookCeph},
		ClusterNamespace_monitoring: {EckOperator, KubePrometheusStack},
		ClusterNamespace_toolkit:    {CertManager, ArgoWorkflows, Kafka, ClusterAutoscaler},
	}
)

func GetInstallOrder() []any {
	return []any{
		Cilium,
		CertManager,
		ClusterIssuer,
		GatewayCrd,
		Metallb,
		IpAddressPool,
		RookCeph,
		EckOperator,
		KubePrometheusStack,
		CertManager,
		Metallb,
		Cilium,
		RookCeph,
		ArgoWorkflows,
		ClusterAutoscaler,
	}
}

func GetUninstallOrder() []any {
	return []any{
		GatewayCrd, ClusterIssuer, IpAddressPool, CephCluster, StorageClass,
		Metallb, Cilium,
		ArgoWorkflows,
		ClusterAutoscaler,
		EckOperator,
		KubePrometheusStack,
		RookCeph,
	}
}

func IsAppName(name any) bool {
	if _, ok := name.(AppName); ok {
		return true
	}
	return false
}

func IsAppCrd(crd any) bool {
	if _, ok := crd.(AppCrd); ok {
		return true
	}
	return false
}

func GetAppNs(appName AppName) (ns string) {
	for ns, apps := range appNsMap {
		if slices.Contains(apps, appName) {
			return ns.String()
		}
	}
	return
}

type AppType struct {
	Id          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type AppRepo struct {
	Id          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Url         string `json:"url,omitempty"`
	IndexPath   string `json:"index_path,omitempty"`
	Description string `json:"description,omitempty"`
	Apps        []*App `json:"apps,omitempty"`
}

type AppVersion struct {
	Id            int64  `json:"id,omitempty"`
	AppId         int64  `json:"app_id,omitempty"`
	Name          string `json:"name,omitempty"`
	Chart         string `json:"chart,omitempty"`
	Version       string `json:"version,omitempty"`
	DefaultConfig string `json:"default_config,omitempty"`
}

type App struct {
	Id          int64         `json:"id,omitempty"`
	Name        string        `json:"name,omitempty"`
	Icon        string        `json:"icon,omitempty"`
	AppTypeId   int64         `json:"app_type_id,omitempty"`
	AppRepoId   int64         `json:"app_repo_id,omitempty"`
	Description string        `json:"description,omitempty"`
	Versions    []*AppVersion `json:"versions,omitempty"`
	Readme      string        `json:"readme,omitempty"`
	Metadata    []byte        `json:"metadata,omitempty"`
}

type AppStatus int32

const (
	AppStatus_PENDING   AppStatus = 0
	AppStatus_COMPLETED AppStatus = 1
	AppStatus_FAILED    AppStatus = 2
)

// Enum value maps for AppStatus.
var (
	AppStatus_name = map[int32]string{
		0: "PENDING",
		1: "COMPLETED",
		2: "FAILED",
	}
	AppStatus_value = map[string]int32{
		"PENDING":   0,
		"COMPLETED": 1,
		"FAILED":    2,
	}
)

func (a AppStatus) String() string {
	return AppStatus_name[int32(a)]
}

func (a AppStatus) Int32() int32 {
	return AppStatus_value[a.String()]
}

// CloudAppStatus defines the observed state of CloudApp.
type CloudAppStatus struct {
	Status         AppStatus   `json:"status,omitempty"`
	LastUpdateTime metav1.Time `json:"last_update_time,omitempty"`
	Message        string      `json:"message,omitempty"`
}

type CloudAppSpec struct {
	App     App     `json:"app,omitempty"`
	AppRepo AppRepo `json:"app_repo,omitempty"`
}

func (a CloudAppSpec) AppIsEmpty() bool {
	return a.App.Id == 0 || a.App.Name == ""
}

func (a CloudAppSpec) AppRepoIsEmpty() bool {
	return a.AppRepo.Id == 0 || a.AppRepo.Name == "" || a.AppRepo.Url == ""
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudApp is the Schema for the cloudapps API.
type CloudApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CloudAppSpec   `json:"spec,omitempty"`
	Status CloudAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudAppList contains a list of CloudApp.
type CloudAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudApp{}, &CloudAppList{})
}
