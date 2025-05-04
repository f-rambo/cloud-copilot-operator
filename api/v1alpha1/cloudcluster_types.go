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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CloudClusterSpec defines the desired state of CloudCluster.
// type CloudClusterSpec struct {
// 	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file

// 	// Foo is an example field of CloudCluster. Edit cloudcluster_types.go to remove/update
// 	Foo string `json:"foo,omitempty"`
// }

// CloudClusterStatus defines the observed state of CloudCluster.
// type CloudClusterStatus struct {
// 	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// 	// Important: Run "make" to regenerate code after modifying this file
// }

type ClusterNamespace int32

const (
	ClusterNamespace_cloudcopilot ClusterNamespace = 0
	ClusterNamespace_networking   ClusterNamespace = 1
	ClusterNamespace_storage      ClusterNamespace = 2
	ClusterNamespace_monitoring   ClusterNamespace = 3
	ClusterNamespace_toolkit      ClusterNamespace = 4
)

// ClusterNamespace to string
func (cn ClusterNamespace) String() string {
	switch cn {
	case ClusterNamespace_cloudcopilot:
		return "cloudcopilot"
	case ClusterNamespace_networking:
		return "networking"
	case ClusterNamespace_storage:
		return "storage"
	case ClusterNamespace_monitoring:
		return "monitoring"
	case ClusterNamespace_toolkit:
		return "toolkit"
	default:
		return "default"
	}
}

type ClusterProvider int32

const (
	ClusterProvider_UNSPECIFIED ClusterProvider = 0
	ClusterProvider_BareMetal   ClusterProvider = 1
	ClusterProvider_Aws         ClusterProvider = 2
	ClusterProvider_AliCloud    ClusterProvider = 3
)

// ClusterProvider to string
func (cp ClusterProvider) String() string {
	switch cp {
	case ClusterProvider_BareMetal:
		return "baremetal"
	case ClusterProvider_Aws:
		return "aws"
	case ClusterProvider_AliCloud:
		return "ali_cloud"
	default:
		return ""
	}
}

type ClusterStatus int32

const (
	ClusterStatus_UNSPECIFIED ClusterStatus = 0
	ClusterStatus_STARTING    ClusterStatus = 1 // Cluster init
	ClusterStatus_RUNNING     ClusterStatus = 2 // Network ready & storage ready & monitoring ready & toolkit ready
	ClusterStatus_STOPPING    ClusterStatus = 3
	ClusterStatus_STOPPED     ClusterStatus = 4
	ClusterStatus_DELETED     ClusterStatus = 5
)

// ClusterStatus to string
func (cs ClusterStatus) String() string {
	switch cs {
	case ClusterStatus_STARTING:
		return "starting"
	case ClusterStatus_RUNNING:
		return "running"
	case ClusterStatus_STOPPING:
		return "stopping"
	case ClusterStatus_STOPPED:
		return "stopped"
	case ClusterStatus_DELETED:
		return "deleted"
	default:
		return ""
	}
}

type ClusterLevel int32

const (
	ClusterLevel_UNSPECIFIED ClusterLevel = 0
	ClusterLevel_BASIC       ClusterLevel = 1
	ClusterLevel_STANDARD    ClusterLevel = 2
	ClusterLevel_ADVANCED    ClusterLevel = 3
)

// ClusterLevel to string
func (cl ClusterLevel) String() string {
	switch cl {
	case ClusterLevel_BASIC:
		return "basic"
	case ClusterLevel_STANDARD:
		return "standard"
	case ClusterLevel_ADVANCED:
		return "advanced"
	default:
		return ""
	}
}

type NodeRole int32

const (
	NodeRole_UNSPECIFIED NodeRole = 0
	NodeRole_MASTER      NodeRole = 1
	NodeRole_WORKER      NodeRole = 2
	NodeRole_EDGE        NodeRole = 3
)

// NodeRole to string
func (nr NodeRole) String() string {
	switch nr {
	case NodeRole_MASTER:
		return "master"
	case NodeRole_WORKER:
		return "worker"
	case NodeRole_EDGE:
		return "edge"
	default:
		return ""
	}
}

type NodeStatus int32

const (
	NodeStatus_UNSPECIFIED   NodeStatus = 0
	NodeStatus_NODE_READY    NodeStatus = 1
	NodeStatus_NODE_FINDING  NodeStatus = 2
	NodeStatus_NODE_CREATING NodeStatus = 3
	NodeStatus_NODE_PENDING  NodeStatus = 4
	NodeStatus_NODE_RUNNING  NodeStatus = 5
	NodeStatus_NODE_DELETING NodeStatus = 6
	NodeStatus_NODE_DELETED  NodeStatus = 7
	NodeStatus_NODE_ERROR    NodeStatus = 8
)

// NodeStatus to string
func (ns NodeStatus) String() string {
	switch ns {
	case NodeStatus_NODE_READY:
		return "node_ready"
	case NodeStatus_NODE_FINDING:
		return "node_finding"
	case NodeStatus_NODE_CREATING:
		return "node_creating"
	case NodeStatus_NODE_PENDING:
		return "node_pending"
	case NodeStatus_NODE_RUNNING:
		return "node_running"
	case NodeStatus_NODE_DELETING:
		return "node_deleting"
	case NodeStatus_NODE_DELETED:
		return "node_deleted"
	case NodeStatus_NODE_ERROR:
		return "node_error"
	default:
		return ""
	}
}

type NodeGroupType int32

const (
	NodeGroupType_UNSPECIFIED      NodeGroupType = 0
	NodeGroupType_NORMAL           NodeGroupType = 1
	NodeGroupType_HIGH_COMPUTATION NodeGroupType = 2
	NodeGroupType_GPU_ACCELERATERD NodeGroupType = 3
	NodeGroupType_HIGH_MEMORY      NodeGroupType = 4
	NodeGroupType_LARGE_HARD_DISK  NodeGroupType = 5
	NodeGroupType_LOAD_DISK        NodeGroupType = 6
)

// NodeGroupType to string
func (ngt NodeGroupType) String() string {
	switch ngt {
	case NodeGroupType_NORMAL:
		return "normal"
	case NodeGroupType_HIGH_COMPUTATION:
		return "high_computation"
	case NodeGroupType_GPU_ACCELERATERD:
		return "gpu_accelerated"
	case NodeGroupType_HIGH_MEMORY:
		return "high_memory"
	case NodeGroupType_LARGE_HARD_DISK:
		return "large_hard_disk"
	case NodeGroupType_LOAD_DISK:
		return "load_disk"
	default:
		return ""
	}
}

type NodeArchType int32

const (
	NodeArchType_UNSPECIFIED NodeArchType = 0
	NodeArchType_AMD64       NodeArchType = 1
	NodeArchType_ARM64       NodeArchType = 2
)

func (n NodeArchType) String() string {
	switch n {
	case NodeArchType_AMD64:
		return "amd64"
	case NodeArchType_ARM64:
		return "arm64"
	default:
		return ""
	}
}

func NodeArchTypeFromString(s string) NodeArchType {
	switch s {
	case "amd64":
		return NodeArchType_AMD64
	case "arm64":
		return NodeArchType_ARM64
	default:
		return 0
	}
}

type NodeGPUSpec int32

const (
	NodeGPUSpec_UNSPECIFIED NodeGPUSpec = 0
	NodeGPUSpec_NVIDIA_A10  NodeGPUSpec = 1
	NodeGPUSpec_NVIDIA_V100 NodeGPUSpec = 2
	NodeGPUSpec_NVIDIA_T4   NodeGPUSpec = 3
	NodeGPUSpec_NVIDIA_P100 NodeGPUSpec = 4
	NodeGPUSpec_NVIDIA_P4   NodeGPUSpec = 5
)

func (n NodeGPUSpec) String() string {
	switch n {
	case NodeGPUSpec_NVIDIA_A10:
		return "nvidia-a10"
	case NodeGPUSpec_NVIDIA_V100:
		return "nvidia-v100"
	case NodeGPUSpec_NVIDIA_T4:
		return "nvidia-t4"
	case NodeGPUSpec_NVIDIA_P100:
		return "nvidia-p100"
	case NodeGPUSpec_NVIDIA_P4:
		return "nvidia-p4"
	default:
		return ""
	}
}

// string to NodeGPUSpec
func NodeGPUSpecFromString(s string) NodeGPUSpec {
	switch s {
	case "nvidia-a10":
		return NodeGPUSpec_NVIDIA_A10
	case "nvidia-v100":
		return NodeGPUSpec_NVIDIA_V100
	case "nvidia-t4":
		return NodeGPUSpec_NVIDIA_T4
	case "nvidia-p100":
		return NodeGPUSpec_NVIDIA_P100
	case "nvidia-p4":
		return NodeGPUSpec_NVIDIA_P4
	default:
		return 0
	}
}

type ResourceType int32

const (
	ResourceType_UNSPECIFIED        ResourceType = 0
	ResourceType_VPC                ResourceType = 1
	ResourceType_SUBNET             ResourceType = 2
	ResourceType_INTERNET_GATEWAY   ResourceType = 3
	ResourceType_NAT_GATEWAY        ResourceType = 4
	ResourceType_ROUTE_TABLE        ResourceType = 5
	ResourceType_SECURITY_GROUP     ResourceType = 6
	ResourceType_LOAD_BALANCER      ResourceType = 7
	ResourceType_ELASTIC_IP         ResourceType = 8
	ResourceType_AVAILABILITY_ZONES ResourceType = 9
	ResourceType_KEY_PAIR           ResourceType = 10
	ResourceType_DATA_DEVICE        ResourceType = 11
	ResourceType_INSTANCE           ResourceType = 12
	ResourceType_REGION             ResourceType = 13
	ResourceType_GATEWAY_CLASS      ResourceType = 14
	ResourceType_STORAGE_CLASS      ResourceType = 15
)

// ResourceType to string
func (rt ResourceType) String() string {
	switch rt {
	case ResourceType_VPC:
		return "vpc"
	case ResourceType_SUBNET:
		return "subnet"
	case ResourceType_INTERNET_GATEWAY:
		return "internet_gateway"
	case ResourceType_NAT_GATEWAY:
		return "nat_gateway"
	case ResourceType_ROUTE_TABLE:
		return "route_table"
	case ResourceType_SECURITY_GROUP:
		return "security_group"
	case ResourceType_LOAD_BALANCER:
		return "load_balancer"
	case ResourceType_ELASTIC_IP:
		return "elastic_ip"
	case ResourceType_AVAILABILITY_ZONES:
		return "availability_zones"
	case ResourceType_KEY_PAIR:
		return "key_pair"
	case ResourceType_DATA_DEVICE:
		return "data_device"
	case ResourceType_INSTANCE:
		return "instance"
	case ResourceType_REGION:
		return "region"
	case ResourceType_GATEWAY_CLASS:
		return "gateway_class"
	case ResourceType_STORAGE_CLASS:
		return "storage_class"
	default:
		return ""
	}
}

type ResourceTypeKeyValue int32

const (
	ResourceTypeKeyValue_UNSPECIFIED      ResourceTypeKeyValue = 0
	ResourceTypeKeyValue_NAME             ResourceTypeKeyValue = 1
	ResourceTypeKeyValue_ACCESS           ResourceTypeKeyValue = 2
	ResourceTypeKeyValue_ZONE_ID          ResourceTypeKeyValue = 3
	ResourceTypeKeyValue_REGION_ID        ResourceTypeKeyValue = 4
	ResourceTypeKeyValue_ACCESS_PRIVATE   ResourceTypeKeyValue = 5
	ResourceTypeKeyValue_ACCESS_PUBLIC    ResourceTypeKeyValue = 6
	ResourceTypeKeyValue_EXTERNAL_GATEWAY ResourceTypeKeyValue = 7
	ResourceTypeKeyValue_INTERNAL_GATEWAY ResourceTypeKeyValue = 8
	ResourceTypeKeyValue_BLOCK_STORAGE    ResourceTypeKeyValue = 9
	ResourceTypeKeyValue_FILE_STORAGE     ResourceTypeKeyValue = 10
	ResourceTypeKeyValue_OBJECT_STORAGE   ResourceTypeKeyValue = 11
)

// ResourceTypeKeyValue to string
func (kv ResourceTypeKeyValue) String() string {
	switch kv {
	case ResourceTypeKeyValue_NAME:
		return "name"
	case ResourceTypeKeyValue_ACCESS:
		return "access"
	case ResourceTypeKeyValue_ZONE_ID:
		return "zone_id"
	case ResourceTypeKeyValue_REGION_ID:
		return "region_id"
	case ResourceTypeKeyValue_ACCESS_PRIVATE:
		return "access_private"
	case ResourceTypeKeyValue_ACCESS_PUBLIC:
		return "access_public"
	case ResourceTypeKeyValue_EXTERNAL_GATEWAY:
		return "external_gateway"
	case ResourceTypeKeyValue_INTERNAL_GATEWAY:
		return "internal_gateway"
	case ResourceTypeKeyValue_BLOCK_STORAGE:
		return "block_storage"
	case ResourceTypeKeyValue_FILE_STORAGE:
		return "file_storage"
	case ResourceTypeKeyValue_OBJECT_STORAGE:
		return "object_storage"
	default:
		return ""
	}
}

type SecurityAccess int32

const (
	SecurityAccess_UNSPECIFIED SecurityAccess = 0
	SecurityAccess_PRIVATE     SecurityAccess = 1
	SecurityAccess_PUBLIC      SecurityAccess = 2 // use slb
)

type NodeErrorType int32

const (
	NodeErrorType_UNSPECIFIED          NodeErrorType = 0
	NodeErrorType_INFRASTRUCTURE_ERROR NodeErrorType = 1
	NodeErrorType_CLUSTER_ERROR        NodeErrorType = 2
)

func (c *Cluster) GetHaveDiskNodes() []*Node {
	var nodes []*Node
	for _, node := range c.Nodes {
		if len(node.Disks) > 0 {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

type Cluster struct {
	Id                int64            `gorm:"column:id;primaryKey;AUTO_INCREMENT" json:"id,omitempty"`
	Name              string           `gorm:"column:name;default:'';NOT NULL" json:"name,omitempty"`
	ApiServerAddress  string           `gorm:"column:api_server_address;default:'';NOT NULL" json:"api_server_address,omitempty"`
	KubernetesVersion string           `gorm:"column:kubernetes_version;default:'';NOT NULL" json:"kubernetes_version,omitempty"`
	ImageRepository   string           `gorm:"column:image_repository;default:'';NOT NULL" json:"image_repository,omitempty"`
	Config            string           `gorm:"column:config;default:'';NOT NULL" json:"config,omitempty"`
	Status            ClusterStatus    `gorm:"column:status;default:0;NOT NULL" json:"status,omitempty"`
	Provider          ClusterProvider  `gorm:"column:provider;default:0;NOT NULL" json:"provider,omitempty"`
	Level             ClusterLevel     `gorm:"column:level;default:0;NOT NULL" json:"level,omitempty"`
	PublicKey         string           `gorm:"column:public_key;default:'';NOT NULL" json:"public_key,omitempty"`
	PrivateKey        string           `gorm:"column:private_key;default:'';NOT NULL" json:"private_key,omitempty"`
	Region            string           `gorm:"column:region;default:'';NOT NULL" json:"region,omitempty"`
	UserId            int64            `gorm:"column:user_id;default:0;NOT NULL" json:"user_id,omitempty"` // action user
	AccessId          string           `gorm:"column:access_id;default:'';NOT NULL" json:"access_id,omitempty"`
	AccessKey         string           `gorm:"column:access_key;default:'';NOT NULL" json:"access_key,omitempty"`
	NodeUsername      string           `gorm:"column:node_username;default:'';NOT NULL" json:"node_username,omitempty"`
	NodeStartIp       string           `gorm:"column:node_start_ip;default:'';NOT NULL" json:"node_start_ip,omitempty"`
	NodeEndIp         string           `gorm:"column:node_end_ip;default:'';NOT NULL" json:"node_end_ip,omitempty"`
	Domain            string           `gorm:"column:domain;default:'';NOT NULL" json:"domain,omitempty"`
	VpcCidr           string           `gorm:"column:vpc_cidr;default:'';NOT NULL" json:"vpc_cidr,omitempty"`
	ServiceCidr       string           `gorm:"column:service_cidr;default:'';NOT NULL" json:"service_cidr,omitempty"`
	PodCidr           string           `gorm:"column:pod_cidr;default:'';NOT NULL" json:"pod_cidr,omitempty"`
	SubnetCidrs       string           `gorm:"column:subnet_cidrs;default:'';NOT NULL" json:"subnet_cidrs,omitempty"` // 多个子网cidr，逗号分隔
	GatewayClass      string           `gorm:"column:gateway_class;default:'';NOT NULL" json:"gateway_class,omitempty"`
	StorageClass      string           `gorm:"column:storage_class;default:'';NOT NULL" json:"storage_class,omitempty"`
	NodeGroups        []*NodeGroup     `gorm:"-" json:"node_groups,omitempty"`
	Nodes             []*Node          `gorm:"-" json:"nodes,omitempty"`
	CloudResources    []*CloudResource `gorm:"-" json:"cloud_resources,omitempty"`
	Securitys         []*Security      `gorm:"-" json:"securitys,omitempty"`
}

type NodeGroup struct {
	Id         string        `gorm:"column:id;primaryKey;NOT NULL" json:"id,omitempty"`
	Name       string        `gorm:"column:name;default:'';NOT NULL" json:"name,omitempty"`
	Type       NodeGroupType `gorm:"column:type;default:0;NOT NULL" json:"type,omitempty"`
	Os         string        `gorm:"column:os;default:'';NOT NULL" json:"os,omitempty"`
	Arch       NodeArchType  `gorm:"column:arch;default:0;NOT NULL" json:"arch,omitempty"`
	Cpu        int32         `gorm:"column:cpu;default:0;NOT NULL" json:"cpu,omitempty"`
	Memory     int32         `gorm:"column:memory;default:0;NOT NULL" json:"memory,omitempty"`
	Gpu        int32         `gorm:"column:gpu;default:0;NOT NULL" json:"gpu,omitempty"`
	GpuSpec    NodeGPUSpec   `gorm:"column:gpu_spec;default:0;NOT NULL" json:"gpu_spec,omitempty"`
	MinSize    int32         `gorm:"column:min_size;default:0;NOT NULL" json:"min_size,omitempty"`
	MaxSize    int32         `gorm:"column:max_size;default:0;NOT NULL" json:"max_size,omitempty"`
	TargetSize int32         `gorm:"column:target_size;default:0;NOT NULL" json:"target_size,omitempty"`
	ClusterId  int64         `gorm:"column:cluster_id;default:0;NOT NULL" json:"cluster_id,omitempty"`
}

type Node struct {
	Id                int64         `gorm:"column:id;primaryKey;AUTO_INCREMENT" json:"id,omitempty"`
	Name              string        `gorm:"column:name;default:'';NOT NULL" json:"name,omitempty"`
	Labels            string        `gorm:"column:labels;default:'';NOT NULL" json:"labels,omitempty"`
	Ip                string        `gorm:"column:ip;default:'';NOT NULL" json:"ip,omitempty"`
	Username          string        `gorm:"column:username;default:'';NOT NULL" json:"username,omitempty"`
	Role              NodeRole      `gorm:"column:role;default:0;NOT NULL" json:"role,omitempty"`
	Status            NodeStatus    `gorm:"column:status;default:0;NOT NULL" json:"status,omitempty"`
	InstanceId        string        `gorm:"column:instance_id;default:'';NOT NULL" json:"instance_id,omitempty"`
	ImageId           string        `gorm:"column:image_id;default:'';NOT NULL" json:"image_id,omitempty"`
	BackupInstanceIds string        `gorm:"column:backup_instance_ids;default:'';NOT NULL" json:"backup_instance_ids,omitempty"`
	InstanceType      string        `gorm:"column:instance_type;default:'';NOT NULL" json:"instance_type,omitempty"`
	Disks             []*Disk       `gorm:"-" json:"disks,omitempty"`
	ClusterId         int64         `gorm:"column:cluster_id;default:0;NOT NULL" json:"cluster_id,omitempty"`
	NodeGroupId       string        `gorm:"column:node_group_id;default:'';NOT NULL" json:"node_group_id,omitempty"`
	NodeInfo          string        `gorm:"column:node_info;default:'';NOT NULL" json:"node_info,omitempty"`
	ErrorType         NodeErrorType `gorm:"column:error_type;default:0;NOT NULL" json:"error_type,omitempty"`
	ErrorMessage      string        `gorm:"column:error_message;default:'';NOT NULL" json:"error_message,omitempty"`
}

type Disk struct {
	Id        string `gorm:"column:id;primaryKey;NOT NULL" json:"id,omitempty"`
	Name      string `gorm:"column:name;default:'';NOT NULL" json:"name,omitempty"`
	Size      int32  `gorm:"column:size;default:0;NOT NULL" json:"size,omitempty"`
	Device    string `gorm:"column:device;default:'';NOT NULL" json:"device,omitempty"`
	NodeId    int64  `gorm:"column:node_id;default:0;NOT NULL" json:"node_id,omitempty"`
	ClusterId int64  `gorm:"column:cluster_id;default:0;NOT NULL" json:"cluster_id,omitempty"`
}

type CloudResource struct {
	Id           string       `gorm:"column:id;primaryKey;NOT NULL" json:"id,omitempty"`
	Name         string       `gorm:"column:name;default:'';NOT NULL" json:"name,omitempty"`
	RefId        string       `gorm:"column:ref_id;default:'';NOT NULL" json:"ref_id,omitempty"`
	AssociatedId string       `gorm:"column:associated_id;default:'';NOT NULL" json:"associated_id,omitempty"`
	Type         ResourceType `gorm:"column:type;default:0;NOT NULL" json:"type,omitempty"`
	Tags         string       `gorm:"column:tags;default:'';NOT NULL" json:"tags,omitempty"`
	Value        string       `gorm:"column:value;default:'';NOT NULL" json:"value,omitempty"`
	ClusterId    int64        `gorm:"column:cluster_id;default:0;NOT NULL" json:"cluster_id,omitempty"`
}

type Security struct {
	Id        string         `gorm:"column:id;primaryKey;NOT NULL" json:"id,omitempty"`
	Name      string         `gorm:"column:name;default:'';NOT NULL" json:"name,omitempty"`
	StartPort int32          `gorm:"column:start_port;default:0;NOT NULL" json:"start_port,omitempty"`
	EndPort   int32          `gorm:"column:end_port;default:0;NOT NULL" json:"end_port,omitempty"`
	Protocol  string         `gorm:"column:protocol;default:'';NOT NULL" json:"protocol,omitempty"`
	IpCidr    string         `gorm:"column:ip_cidr;default:'';NOT NULL" json:"ip_cidr,omitempty"`
	Access    SecurityAccess `gorm:"column:access;default:0;NOT NULL" json:"access,omitempty"`
	ClusterId int64          `gorm:"column:cluster_id;default:0;NOT NULL" json:"cluster_id,omitempty"`
}

// string to float32
func StringToFloat32(s string) float32 {
	var f float32
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return 0
	}
	return f
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CloudCluster is the Schema for the cloudclusters API.
type CloudCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Cluster       `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CloudClusterList contains a list of CloudCluster.
type CloudClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudCluster{}, &CloudClusterList{})
}
