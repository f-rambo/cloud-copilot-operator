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
	"path/filepath"
	"strings"

	cloudcopilotv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	"github.com/f-rambo/cloud-copilot/operator/utils"
	"github.com/google/uuid"
	yaml "gopkg.in/yaml.v3"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
	helmrelease "helm.sh/helm/v3/pkg/release"
	coreV1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudAppReleaseReconciler reconciles a CloudAppRelease object
type CloudAppReleaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudappreleases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudappreleases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudappreleases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CloudAppRelease object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *CloudAppReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	cloudAppRelease := &cloudcopilotv1alpha1.CloudAppRelease{}
	err := r.Get(ctx, req.NamespacedName, cloudAppRelease)
	if err != nil {
		if apierror.IsNotFound(err) {
			err = DeleteAppRelease(ctx, &cloudcopilotv1alpha1.AppRelease{
				ReleaseName: req.NamespacedName.Name,
				Namespace:   req.NamespacedName.Namespace,
				Wait:        true,
			})
			if err != nil {
				log.Error(err, "DeleteAppRelease failed")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if cloudAppRelease.Status.Status != cloudcopilotv1alpha1.AppReleaseSatus_PENDING && cloudAppRelease.Status.Status != cloudcopilotv1alpha1.AppReleaseSatus_UNSPECIFIED {
		return ctrl.Result{}, nil
	}
	cloudAppRelease.Spec.Wait = true
	err = AppRelease(ctx, &cloudAppRelease.Spec)
	if err != nil {
		log.Error(err, "AppRelease failed")
		cloudAppRelease.Status.Status = cloudcopilotv1alpha1.AppReleaseSatus_FAILED
		err = r.Update(ctx, cloudAppRelease)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	err = GetAppReleaseResources(ctx, cloudAppRelease)
	if err != nil {
		log.Error(err, "GetAppReleaseResources failed")
		return ctrl.Result{}, err
	}
	for _, item := range cloudAppRelease.Spec.Resources {
		err = ReloadAppReleaseResource(ctx, item)
		if err != nil {
			log.Error(err, "ReloadAppReleaseResource failed")
			return ctrl.Result{}, err
		}
	}
	err = GetAppReleaseStatus(ctx, &cloudAppRelease.Spec)
	if err != nil {
		log.Error(err, "GetAppReleaseStatus failed")
		return ctrl.Result{}, err
	}
	err = r.Update(ctx, cloudAppRelease)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudAppReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcopilotv1alpha1.CloudAppRelease{}).
		Named("cloudapprelease").
		Complete(r)
}

func AppRelease(ctx context.Context, appRelease *cloudcopilotv1alpha1.AppRelease) error {
	helmPkg, err := utils.NewHelmPkg(appRelease.Namespace)
	if err != nil {
		return err
	}
	install, err := helmPkg.NewInstall()
	if err != nil {
		return err
	}
	if appRelease.RepoName != "" && appRelease.Chart == "" {
		appsPath := AppPath
		pullClient, getErr := helmPkg.NewPull()
		if getErr != nil {
			return getErr
		}
		err = helmPkg.RunPull(pullClient, appsPath, filepath.Join(
			appRelease.RepoName,
			appRelease.AppName), appRelease.AppVersion)
		if err != nil {
			return err
		}
		appRelease.Chart, err = utils.FindMatchingFile(appsPath, appRelease.AppName)
		if err != nil {
			return err
		}
	}
	charInfo, err := utils.GetLocalChartInfo(appRelease.Chart)
	if err != nil {
		return err
	}
	appRelease.AppVersion = charInfo.Version
	install.ReleaseName = appRelease.ReleaseName
	install.Namespace = appRelease.Namespace
	install.CreateNamespace = true
	install.GenerateName = true
	install.Version = appRelease.AppVersion
	install.DryRun = appRelease.Dryrun
	install.Atomic = appRelease.Atomic
	install.Wait = appRelease.Wait
	release, err := helmPkg.RunInstall(ctx, install, appRelease.Chart, &helmValues.Options{
		ValueFiles: []string{appRelease.ConfigFile},
		Values:     []string{appRelease.Config},
	})
	appRelease.Logs = helmPkg.GetLogs()
	if err != nil {
		return err
	}
	if release != nil {
		appRelease.ReleaseName = release.Name
		if release.Info != nil {
			appRelease.Notes = release.Info.Notes
		}
		return nil
	}
	return nil
}

func GetAppReleaseStatus(ctx context.Context, appRelease *cloudcopilotv1alpha1.AppRelease) error {
	helmPkg, err := utils.NewHelmPkg(appRelease.Namespace)
	if err != nil {
		return err
	}
	list, err := helmPkg.NewList()
	if err != nil {
		return err
	}
	data, err := helmPkg.RunList(list)
	if err != nil {
		return err
	}
	appRelease.Status = cloudcopilotv1alpha1.AppReleaseSatus_UNSPECIFIED
	for _, r := range data {
		if r.Name == appRelease.ReleaseName && r.Info != nil {
			if r.Info.Status == helmrelease.StatusDeployed {
				appRelease.Status = cloudcopilotv1alpha1.AppReleaseSatus_RUNNING
			}
			if r.Info.Status == helmrelease.StatusPendingInstall || r.Info.Status == helmrelease.StatusPendingUpgrade || r.Info.Status == helmrelease.StatusPendingRollback {
				appRelease.Status = cloudcopilotv1alpha1.AppReleaseSatus_PENDING
			}
			break
		}
	}
	return nil
}

func DeleteAppRelease(ctx context.Context, appRelease *cloudcopilotv1alpha1.AppRelease) error {
	helmPkg, err := utils.NewHelmPkg(appRelease.Namespace)
	if err != nil {
		return err
	}
	uninstall, err := helmPkg.NewUninstall()
	if err != nil {
		return err
	}
	uninstall.KeepHistory = false
	uninstall.DryRun = appRelease.Dryrun
	uninstall.Wait = appRelease.Wait
	resp, err := helmPkg.RunUninstall(uninstall, appRelease.ReleaseName)
	if err != nil {
		return err
	}
	appRelease.Logs = helmPkg.GetLogs()
	appRelease.Notes = resp.Info
	return nil
}

func GetAppReleaseResources(ctx context.Context, cloudAppRelease *cloudcopilotv1alpha1.CloudAppRelease) error {
	helmPkg, err := utils.NewHelmPkg(cloudAppRelease.Spec.Namespace)
	if err != nil {
		return err
	}
	releaseInfo, err := helmPkg.GetReleaseInfo(cloudAppRelease.Spec.ReleaseName)
	if err != nil {
		return err
	}
	if releaseInfo == nil || releaseInfo.Manifest == "" {
		return nil
	}
	clusterClient, err := utils.GetKubeClientByKubeConfig()
	if err != nil {
		return err
	}
	events, err := clusterClient.CoreV1().Events(cloudAppRelease.Spec.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	resources := strings.Split(releaseInfo.Manifest, "---")
	appReleaseResources := make([]*cloudcopilotv1alpha1.AppReleaseResource, 0)
	for _, resource := range resources {
		if resource == "" {
			continue
		}
		obj := unstructured.Unstructured{}
		err = yaml.Unmarshal([]byte(resource), &obj.Object)
		if err != nil {
			return err
		}
		lableStr := ""
		for k, v := range obj.GetLabels() {
			lableStr += fmt.Sprintf("%s=%s,", k, v)
		}
		lableStr = strings.TrimRight(lableStr, ",")
		appReleaseResource := &cloudcopilotv1alpha1.AppReleaseResource{
			Id:        uuid.NewString(),
			ReleaseId: cloudAppRelease.Spec.Id,
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Kind:      obj.GetKind(),
			Lables:    lableStr,
			Manifest:  resource,
			Status:    cloudcopilotv1alpha1.AppReleaseResourceStatus_SUCCESSFUL,
		}
		eventsStr := ""
		for _, event := range events.Items {
			if event.InvolvedObject.Name == obj.GetName() {
				if event.Type != "Normal" {
					appReleaseResource.Status = cloudcopilotv1alpha1.AppReleaseResourceStatus_UNHEALTHY
				}
				eventsStr += fmt.Sprintf("- Type: %s, Reason: %s, Message: %s\n", event.Type, event.Reason, event.Message)
			}
		}
		appReleaseResource.Events = eventsStr
		appReleaseResources = append(appReleaseResources, appReleaseResource)

		podLableSelector := ""
		if strings.EqualFold(appReleaseResource.Kind, cloudcopilotv1alpha1.Deployment) {
			deploymentResource, err := clusterClient.AppsV1().Deployments(appReleaseResource.Namespace).Get(ctx, appReleaseResource.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			for k, v := range deploymentResource.Spec.Selector.MatchLabels {
				podLableSelector += fmt.Sprintf("%s=%s,", k, v)
			}
			podLableSelector = strings.TrimRight(podLableSelector, ",")
		}
		if strings.EqualFold(appReleaseResource.Kind, cloudcopilotv1alpha1.StatefulSet) {
			statefulSetResource, err := clusterClient.AppsV1().StatefulSets(appReleaseResource.Namespace).Get(ctx, appReleaseResource.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			for k, v := range statefulSetResource.Spec.Selector.MatchLabels {
				podLableSelector += fmt.Sprintf("%s=%s,", k, v)
			}
			podLableSelector = strings.TrimRight(podLableSelector, ",")
		}
		if strings.EqualFold(appReleaseResource.Kind, cloudcopilotv1alpha1.DaemonSet) {
			daemonSetResource, err := clusterClient.AppsV1().DaemonSets(appReleaseResource.Namespace).Get(ctx, appReleaseResource.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			for k, v := range daemonSetResource.Spec.Selector.MatchLabels {
				podLableSelector += fmt.Sprintf("%s=%s,", k, v)
			}
			podLableSelector = strings.TrimRight(podLableSelector, ",")
		}
		if podLableSelector != "" {
			podResources, err := clusterClient.CoreV1().Pods(appReleaseResource.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: podLableSelector,
			})
			if err != nil {
				return err
			}
			for _, pod := range podResources.Items {
				lableStr := ""
				for k, v := range pod.Labels {
					lableStr += fmt.Sprintf("%s=%s,", k, v)
				}
				lableStr = strings.TrimRight(lableStr, ",")
				appReleaseResource := &cloudcopilotv1alpha1.AppReleaseResource{
					Id:        uuid.NewString(),
					ReleaseId: cloudAppRelease.Spec.Id,
					Name:      fmt.Sprintf("%s-%s-%s", obj.GetKind(), obj.GetName(), pod.Name),
					Namespace: pod.Namespace,
					Kind:      cloudcopilotv1alpha1.Pod,
					Lables:    lableStr,
					Status:    cloudcopilotv1alpha1.AppReleaseResourceStatus_SUCCESSFUL,
				}
				eventsStr := ""
				for _, event := range events.Items {
					if string(event.InvolvedObject.UID) == string(pod.UID) {
						if event.Type != coreV1.EventTypeNormal {
							appReleaseResource.Status = cloudcopilotv1alpha1.AppReleaseResourceStatus_UNHEALTHY
						}
						eventsStr += fmt.Sprintf("- Type: %s, Reason: %s, Message: %s\n", event.Type, event.Reason, event.Message)
					}
				}
				if pod.Status.Phase != coreV1.PodRunning {
					appReleaseResource.Status = cloudcopilotv1alpha1.AppReleaseResourceStatus_UNHEALTHY
				}
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if !containerStatus.Ready {
						appReleaseResource.Status = cloudcopilotv1alpha1.AppReleaseResourceStatus_UNHEALTHY
						break
					}
				}
				appReleaseResource.Events = eventsStr
				appReleaseResources = append(appReleaseResources, appReleaseResource)
			}
		}
	}
	cloudAppRelease.Spec.Resources = appReleaseResources
	return nil
}

func ReloadAppReleaseResource(ctx context.Context, appReleaseResource *cloudcopilotv1alpha1.AppReleaseResource) error {
	clusterClient, err := utils.GetKubeClientByKubeConfig()
	if err != nil {
		return err
	}
	if strings.EqualFold(appReleaseResource.Kind, cloudcopilotv1alpha1.Pod) {
		err = clusterClient.CoreV1().Pods(appReleaseResource.Namespace).Delete(ctx, appReleaseResource.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return nil
	}
	podLableSelector := ""
	if strings.EqualFold(appReleaseResource.Kind, cloudcopilotv1alpha1.Deployment) {
		deploymentResource, err2 := clusterClient.AppsV1().Deployments(appReleaseResource.Namespace).Get(ctx, appReleaseResource.Name, metav1.GetOptions{})
		if err2 != nil {
			return err2
		}
		for k, v := range deploymentResource.Spec.Selector.MatchLabels {
			podLableSelector += fmt.Sprintf("%s=%s,", k, v)
		}
		podLableSelector = strings.TrimRight(podLableSelector, ",")
	}
	if strings.EqualFold(appReleaseResource.Kind, cloudcopilotv1alpha1.StatefulSet) {
		statefulSetResource, err2 := clusterClient.AppsV1().StatefulSets(appReleaseResource.Namespace).Get(ctx, appReleaseResource.Name, metav1.GetOptions{})
		if err2 != nil {
			return err2
		}
		for k, v := range statefulSetResource.Spec.Selector.MatchLabels {
			podLableSelector += fmt.Sprintf("%s=%s,", k, v)
		}
		podLableSelector = strings.TrimRight(podLableSelector, ",")
	}
	if strings.EqualFold(appReleaseResource.Kind, cloudcopilotv1alpha1.DaemonSet) {
		daemonSetResource, err2 := clusterClient.AppsV1().DaemonSets(appReleaseResource.Namespace).Get(ctx, appReleaseResource.Name, metav1.GetOptions{})
		if err2 != nil {
			return err2
		}
		for k, v := range daemonSetResource.Spec.Selector.MatchLabels {
			podLableSelector += fmt.Sprintf("%s=%s,", k, v)
		}
		podLableSelector = strings.TrimRight(podLableSelector, ",")
	}
	if podLableSelector == "" {
		return nil
	}
	podResources, err := clusterClient.CoreV1().Pods(appReleaseResource.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: podLableSelector,
	})
	if err != nil {
		return err
	}
	for _, pod := range podResources.Items {
		err = clusterClient.CoreV1().Pods(appReleaseResource.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
