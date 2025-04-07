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

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudcopilotv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	"github.com/f-rambo/cloud-copilot/operator/utils"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sType "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudClusterReconciler reconciles a CloudCluster object
type CloudClusterReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	DynamicInterface dynamic.Interface
}

// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CloudCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *CloudClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	cloudcluster := &cloudcopilotv1alpha1.CloudCluster{}
	err := r.Get(ctx, req.NamespacedName, cloudcluster)
	if err != nil {
		if errors.IsNotFound(err) {
			uninstallOrder := cloudcopilotv1alpha1.GetUninstallOrder()
			for _, item := range uninstallOrder {
				if cloudcopilotv1alpha1.IsAppName(item) {
					err = r.uninstallApp(ctx, item.(cloudcopilotv1alpha1.AppName))
					if err != nil {
						return ctrl.Result{}, err
					}
				}
				if cloudcopilotv1alpha1.IsAppCrd(item) {
					err = r.uninstallCrd(ctx, item.(cloudcopilotv1alpha1.AppCrd), &cloudcluster.Spec)
					if err != nil {
						return ctrl.Result{}, err
					}
				}
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	installItems := cloudcopilotv1alpha1.GetInstallOrder()
	for _, item := range installItems {
		if cloudcopilotv1alpha1.IsAppName(item) {
			err = r.installApp(ctx, item.(cloudcopilotv1alpha1.AppName), &cloudcluster.Spec)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		if cloudcopilotv1alpha1.IsAppCrd(item) {
			err = r.installCrd(ctx, item.(cloudcopilotv1alpha1.AppCrd), &cloudcluster.Spec)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	err = r.HandlerNodes(ctx, &cloudcluster.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// instlal app
func (r *CloudClusterReconciler) installApp(ctx context.Context, appName cloudcopilotv1alpha1.AppName, cluster *cloudcopilotv1alpha1.Cluster) error {
	appRelease := &cloudcopilotv1alpha1.AppRelease{
		ReleaseName: appName.String(),
		AppName:     appName.String(),
		Namespace:   cloudcopilotv1alpha1.GetAppNs(appName),
		Wait:        true,
	}
	err := GetAppReleaseStatus(ctx, appRelease)
	if err != nil {
		return err
	}
	if appRelease.Status == cloudcopilotv1alpha1.AppReleaseSatus_RUNNING || appRelease.Status == cloudcopilotv1alpha1.AppReleaseSatus_PENDING {
		return nil
	}
	appPath, err := utils.FindMatchingFile(AppPath, appName.String())
	if err != nil {
		return err
	}
	appConfigPath, err := utils.FindMatchingFile(AppConfigDir, appName.String())
	if err != nil {
		return err
	}
	tempConfigPath, err := utils.TransferredMeaning(cluster, appConfigPath)
	if err != nil {
		return err
	}
	appRelease.Chart = appPath
	appRelease.ConfigFile = tempConfigPath
	err = AppRelease(ctx, appRelease)
	if err != nil {
		return err
	}
	return nil
}

// uninstall app
func (r *CloudClusterReconciler) uninstallApp(ctx context.Context, appName cloudcopilotv1alpha1.AppName) error {
	appRelease := &cloudcopilotv1alpha1.AppRelease{
		ReleaseName: appName.String(),
		AppName:     appName.String(),
		Namespace:   cloudcopilotv1alpha1.GetAppNs(appName),
		Wait:        true,
	}
	err := DeleteAppRelease(ctx, appRelease)
	if err != nil {
		return err
	}
	return nil
}

// instlal crd
func (r *CloudClusterReconciler) installCrd(ctx context.Context, crdName cloudcopilotv1alpha1.AppCrd, cluster *cloudcopilotv1alpha1.Cluster) error {
	crdFile, err := utils.FindMatchingFile(ComponentDir, crdName.String())
	if err != nil {
		return err
	}
	tempFile, err := utils.TransferredMeaning(cluster, crdFile)
	if err != nil {
		return err
	}
	unstructuredList, err := utils.ParseYAML(tempFile)
	if err != nil {
		return err
	}
	for _, item := range unstructuredList.Items {
		namespacedName := k8sType.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()}
		r.Get(ctx, namespacedName, nil)
		if _, err := getResource(ctx, r.DynamicInterface, &item); err != nil {
			if errors.IsNotFound(err) {
				err = createResource(ctx, r.DynamicInterface, &item)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			err = updateResource(ctx, r.DynamicInterface, &item)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// uninstall crd
func (r *CloudClusterReconciler) uninstallCrd(ctx context.Context, crdName cloudcopilotv1alpha1.AppCrd, cluster *cloudcopilotv1alpha1.Cluster) error {
	crdFile, err := utils.FindMatchingFile(ComponentDir, crdName.String())
	if err != nil {
		return err
	}
	tempFile, err := utils.TransferredMeaning(cluster, crdFile)
	if err != nil {
		return err
	}
	unstructuredList, err := utils.ParseYAML(tempFile)
	if err != nil {
		return err
	}
	for _, item := range unstructuredList.Items {
		namespacedName := k8sType.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()}
		r.Get(ctx, namespacedName, nil)
		if _, err := getResource(ctx, r.DynamicInterface, &item); err != nil {
			if errors.IsNotFound(err) {
				return nil
			} else {
				return err
			}
		} else {
			err = deleteResource(ctx, r.DynamicInterface, &item)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *CloudClusterReconciler) HandlerNodes(ctx context.Context, cluster *cloudcopilotv1alpha1.Cluster) error {
	clientset, err := utils.GetKubeClientByKubeConfig()
	if err != nil {
		return err
	}
	for _, node := range cluster.Nodes {
		if node.Status != cloudcopilotv1alpha1.NodeStatus_NODE_DELETING {
			continue
		}

		pods, getErr := r.getPodsOnNode(ctx, clientset, node.Name)
		if getErr != nil {
			return fmt.Errorf("failed to get pods on node %s: %v", node.Name, getErr)
		}

		if err1 := r.evictPods(ctx, clientset, pods); err1 != nil {
			return fmt.Errorf("failed to evict pods: %v", err1)
		}

		if err2 := r.waitForPodsToBeDeleted(ctx, clientset, pods, 5*time.Minute); err2 != nil {
			return fmt.Errorf("timeout waiting for pods to be deleted: %v", err2)
		}

		err = clientset.CoreV1().Nodes().Delete(ctx, node.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	err = r.getNodes(ctx, clientset, cluster)
	if err != nil {
		return err
	}
	return nil
}

func (r *CloudClusterReconciler) getNodes(ctx context.Context, clientSet *kubernetes.Clientset, cluster *cloudcopilotv1alpha1.Cluster) error {
	nodeRes, err := clientSet.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodeRes.Items {
		n := &cloudcopilotv1alpha1.Node{}
		clusterNodeIndex := -1
		for index, v := range cluster.Nodes {
			if v.Status == cloudcopilotv1alpha1.NodeStatus_NODE_DELETING {
				continue
			}
			if v.Name == node.Name {
				n = v
				clusterNodeIndex = index
				break
			}
		}
		for _, v := range node.Status.Addresses {
			if v.Address == "" {
				continue
			}
			if v.Type == corev1.NodeInternalIP {
				n.Ip = v.Address
			}
		}
		for _, v := range node.Status.Conditions {
			switch v.Type {
			case corev1.NodeReady:
				if v.Status == corev1.ConditionFalse {
					n.Status = cloudcopilotv1alpha1.NodeStatus_NODE_ERROR
					n.ErrorType = cloudcopilotv1alpha1.NodeErrorType_CLUSTER_ERROR
					n.ErrorMessage += fmt.Sprintf("Reason: %s, Message: %s", v.Reason, v.Message)
				}
			case corev1.NodeMemoryPressure:
				if v.Status == corev1.ConditionTrue {
					n.Status = cloudcopilotv1alpha1.NodeStatus_NODE_ERROR
					n.ErrorType = cloudcopilotv1alpha1.NodeErrorType_CLUSTER_ERROR
					n.ErrorMessage += fmt.Sprintf("Reason: %s, Message: %s", v.Reason, v.Message)
				}
			case corev1.NodeDiskPressure:
				if v.Status == corev1.ConditionTrue {
					n.Status = cloudcopilotv1alpha1.NodeStatus_NODE_ERROR
					n.ErrorType = cloudcopilotv1alpha1.NodeErrorType_CLUSTER_ERROR
					n.ErrorMessage += fmt.Sprintf("Reason: %s, Message: %s", v.Reason, v.Message)
				}
			case corev1.NodePIDPressure:
				if v.Status == corev1.ConditionTrue {
					n.Status = cloudcopilotv1alpha1.NodeStatus_NODE_ERROR
					n.ErrorType = cloudcopilotv1alpha1.NodeErrorType_CLUSTER_ERROR
					n.ErrorMessage += fmt.Sprintf("Reason: %s, Message: %s", v.Reason, v.Message)
				}
			case corev1.NodeNetworkUnavailable:
				if v.Status == corev1.ConditionTrue {
					n.Status = cloudcopilotv1alpha1.NodeStatus_NODE_ERROR
					n.ErrorType = cloudcopilotv1alpha1.NodeErrorType_CLUSTER_ERROR
					n.ErrorMessage += fmt.Sprintf("Reason: %s, Message: %s", v.Reason, v.Message)
				}
			default:
				n.Status = cloudcopilotv1alpha1.NodeStatus_NODE_ERROR
				n.ErrorType = cloudcopilotv1alpha1.NodeErrorType_CLUSTER_ERROR
				n.ErrorMessage += fmt.Sprintf("Reason: %s, Message: %s", corev1.ConditionUnknown, v.Message)
			}
		}
		for k, v := range node.Status.Capacity {
			gi, _ := toGiMi(&v)
			if k == corev1.ResourceCPU {
				n.NodeInfo += fmt.Sprintf("CPU(gi): %d ", gi)
			}
			if k == corev1.ResourceMemory {
				n.NodeInfo += fmt.Sprintf("Memory(gi): %d ", gi)
			}
			if k == corev1.ResourceEphemeralStorage {
				n.NodeInfo += fmt.Sprintf("Disk(gi): %d ", gi)
			}
			if strings.Contains(strings.ToUpper(k.String()), "GPU") {
				n.NodeInfo += fmt.Sprintf("GPU(%s): %d ", k.String(), gi)
			}
		}
		if clusterNodeIndex == -1 {
			cluster.Nodes = append(cluster.Nodes, n)
		} else {
			cluster.Nodes[clusterNodeIndex] = n
		}
	}
	return nil
}

func toGiMi(q *resource.Quantity) (int32, int32) {
	var gi, mi float64
	if q.Format == resource.BinarySI {
		gi = float64(q.Value()) / (1 << 30) // 1 Gi = 2^30 bytes
		mi = float64(q.Value()) / (1 << 20) // 1 Mi = 2^20 bytes
	}
	if q.Format == resource.DecimalSI {
		gi = float64(q.Value())
		mi = float64(q.Value()) * 1024
	}
	return int32(gi), int32(mi)
}

func (r *CloudClusterReconciler) getPodsOnNode(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) ([]corev1.Pod, error) {
	podList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, err
	}

	var podsToEvict []corev1.Pod
	for _, pod := range podList.Items {
		if r.isMirrorPod(&pod) || r.isDaemonSetPod(&pod) {
			continue
		}
		podsToEvict = append(podsToEvict, pod)
	}
	return podsToEvict, nil
}

func (r *CloudClusterReconciler) isMirrorPod(pod *corev1.Pod) bool {
	_, exists := pod.Annotations[corev1.MirrorPodAnnotationKey]
	return exists
}

func (r *CloudClusterReconciler) isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

func (r *CloudClusterReconciler) evictPods(ctx context.Context, clientset *kubernetes.Clientset, pods []corev1.Pod) error {
	for _, pod := range pods {
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}
		err := clientset.PolicyV1().Evictions(eviction.Namespace).Evict(ctx, eviction)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CloudClusterReconciler) waitForPodsToBeDeleted(ctx context.Context, clientset *kubernetes.Clientset, pods []corev1.Pod, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, true, func(ctx context.Context) (done bool, err error) {
		for _, pod := range pods {
			_, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if err == nil {
				return false, nil
			}
		}
		return true, nil
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcopilotv1alpha1.CloudCluster{}).
		Named("cloudcluster").
		Complete(r)
}
