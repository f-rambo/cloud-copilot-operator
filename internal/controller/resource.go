package controller

import (
	"fmt"

	operatorv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const MiB = 1024 * 1024 // 1 MiB = 2^20 字节

func NewDeployment(cloudService *operatorv1alpha1.CloudService) *appv1.Deployment {
	resourceList := corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(cloudService.Spec.RequestCPU), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(cloudService.Spec.RequestMemory)*MiB, resource.BinarySI),
	}
	if cloudService.Spec.RequestGPU > 0 {
		resourceList[corev1.ResourceName("nvidia.com/gpu")] = *resource.NewQuantity(int64(cloudService.Spec.RequestGPU), resource.DecimalSI)
	}
	limits := corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(cloudService.Spec.LimitCPU), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(cloudService.Spec.LimitMemory)*MiB, resource.BinarySI),
	}
	if cloudService.Spec.LimitGPU > 0 {
		limits[corev1.ResourceName("nvidia.com/gpu")] = *resource.NewQuantity(int64(cloudService.Spec.LimitGPU), resource.DecimalSI)
	}
	ports := make([]corev1.ContainerPort, 0)
	for _, v := range cloudService.Spec.Ports {
		ports = append(ports, corev1.ContainerPort{
			Name:          v.Name,
			ContainerPort: v.ContainerPort,
			Protocol:      corev1.Protocol(v.Protocol),
		})
	}
	var configfile string
	for file := range cloudService.Spec.Config {
		configfile = file
		break
	}
	container := corev1.Container{
		Name:  cloudService.Name,
		Image: cloudService.Spec.Image,
		Resources: corev1.ResourceRequirements{
			Requests: resourceList,
			Limits:   limits,
		},
		Ports: ports,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      getServiceConfigVolume(cloudService.Name),
				MountPath: cloudService.Spec.ConfigPath,
				SubPath:   configfile,
			},
		},
	}
	configVolume := corev1.Volume{
		Name: getServiceConfigVolume(cloudService.Name),
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cloudService.Name,
				},
			},
		},
	}
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cloudService.Name,
			Namespace: cloudService.Namespace,
			Labels:    cloudService.Labels,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &cloudService.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: cloudService.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cloudService.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes:    []corev1.Volume{configVolume},
				},
			},
		},
	}
}

func NewStatefulSet(cloudService *operatorv1alpha1.CloudService) *appv1.StatefulSet {
	resourceList := corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(cloudService.Spec.RequestCPU), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(cloudService.Spec.RequestMemory)*MiB, resource.BinarySI),
	}
	if cloudService.Spec.RequestGPU > 0 {
		resourceList[corev1.ResourceName("nvidia.com/gpu")] = *resource.NewQuantity(int64(cloudService.Spec.RequestGPU), resource.DecimalSI)
	}
	limits := corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewQuantity(int64(cloudService.Spec.LimitCPU), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(cloudService.Spec.LimitMemory)*MiB, resource.BinarySI),
	}
	if cloudService.Spec.LimitGPU > 0 {
		limits[corev1.ResourceName("nvidia.com/gpu")] = *resource.NewQuantity(int64(cloudService.Spec.LimitGPU), resource.DecimalSI)
	}

	ports := make([]corev1.ContainerPort, 0)
	for _, v := range cloudService.Spec.Ports {
		ports = append(ports, corev1.ContainerPort{
			Name:          v.Name,
			ContainerPort: v.ContainerPort,
			Protocol:      corev1.Protocol(v.Protocol),
		})
	}

	volumeMounts := make([]corev1.VolumeMount, 0)
	var configfile string
	for file := range cloudService.Spec.Config {
		configfile = file
		break
	}
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      getServiceConfigVolume(cloudService.Name),
		MountPath: cloudService.Spec.ConfigPath,
		SubPath:   configfile,
	})
	for _, v := range cloudService.Spec.Volumes {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      v.Name,
			MountPath: v.Path,
		})
	}

	container := corev1.Container{
		Name:  cloudService.Name,
		Image: cloudService.Spec.Image,
		Resources: corev1.ResourceRequirements{
			Requests: resourceList,
			Limits:   limits,
		},
		Ports:        ports,
		VolumeMounts: volumeMounts,
	}

	configVolume := corev1.Volume{
		Name: getServiceConfigVolume(cloudService.Name),
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cloudService.Name,
				},
			},
		},
	}

	volumeClaimTemplates := make([]corev1.PersistentVolumeClaim, 0)
	for _, v := range cloudService.Spec.Volumes {
		volumeClaimTemplates = append(volumeClaimTemplates, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: v.Name,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: *resource.NewQuantity(int64(v.Storage), resource.BinarySI),
					},
				},
			},
		})
	}

	return &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cloudService.Name,
			Namespace: cloudService.Namespace,
			Labels:    cloudService.Labels,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		Spec: appv1.StatefulSetSpec{
			Replicas: &cloudService.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: cloudService.Labels,
			},
			ServiceName: cloudService.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cloudService.Name,
					Namespace: cloudService.Namespace,
					Labels:    cloudService.Labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes:    []corev1.Volume{configVolume},
				},
			},
			VolumeClaimTemplates: volumeClaimTemplates,
		},
	}
}

func NewIngress(cloudService *operatorv1alpha1.CloudService) *netv1.Ingress {
	pathType := netv1.PathTypePrefix
	ingressClassName := "traefik"
	rules := make([]netv1.IngressRule, 0)
	paths := make([]netv1.HTTPIngressPath, 0)
	for _, v := range cloudService.Spec.Ports {
		paths = append(paths, netv1.HTTPIngressPath{
			Path:     v.IngressPath,
			PathType: &pathType,
			Backend: netv1.IngressBackend{
				Service: &netv1.IngressServiceBackend{
					Name: cloudService.Name,
					Port: netv1.ServiceBackendPort{
						Number: v.ContainerPort,
					},
				},
			},
		})
	}
	rules = append(rules, netv1.IngressRule{
		Host: cloudService.Name,
		IngressRuleValue: netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	})
	return &netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cloudService.Name,
			Namespace: cloudService.Namespace,
			Labels:    cloudService.Labels,
		},
		Spec: netv1.IngressSpec{
			Rules:            rules,
			IngressClassName: &ingressClassName,
		},
	}
}

func NewService(cloudService *operatorv1alpha1.CloudService) *corev1.Service {
	ports := make([]corev1.ServicePort, 0)
	for _, v := range cloudService.Spec.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       v.Name,
			Protocol:   corev1.Protocol(v.Protocol),
			Port:       v.ContainerPort,
			TargetPort: intstr.FromInt(int(v.ContainerPort)),
		})
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cloudService.Name,
			Namespace: cloudService.Namespace,
			Labels:    cloudService.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: cloudService.Labels,
			Ports:    ports,
		},
	}
}

func NewConfigMap(cloudService *operatorv1alpha1.CloudService) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cloudService.Name,
			Namespace: cloudService.Namespace,
			Labels:    cloudService.Labels,
		},
		Data: cloudService.Spec.Config,
	}
}

func getServiceConfigVolume(serviceName string) string {
	return fmt.Sprintf("%s-config", serviceName)
}
