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
	"slices"

	cloudServicev1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	"github.com/f-rambo/cloud-copilot/operator/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	k8sType "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudServiceReconciler reconciles a CloudService object
type CloudServiceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	DynamicInterface dynamic.Interface
	ComponentDir     string
}

const FinalizerName = "finalizer.cloud-copilot.operator.io"

// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CloudService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *CloudServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Get CR
	cloudService := &cloudServicev1alpha1.CloudService{}
	err := r.Get(ctx, req.NamespacedName, cloudService)
	if err != nil {
		if errors.IsNotFound(err) {
			// CR no longer exists, remove finalizer if it exists
			log.Info("CloudService deleted, cleaning up resources", "namespace", req.Namespace, "name", req.Name)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CloudService")
		return ctrl.Result{}, err
	}

	if cloudService.Spec.CloudServiceType != cloudServicev1alpha1.CloudServiceTypeGrpcServer && cloudService.Spec.CloudServiceType != cloudServicev1alpha1.CloudServiceTypeHttpServer {
		cloudService.Spec.CloudServiceType = cloudServicev1alpha1.CloudServiceTypeHttpServer
	}

	// Add finalizer if it doesn't exist
	if !slices.Contains(cloudService.Finalizers, FinalizerName) {
		cloudService.Finalizers = append(cloudService.Finalizers, FinalizerName)
		if err := r.Update(ctx, cloudService); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Handle deletion
	if cloudService.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, cloudService)
	}

	// Create or update resources
	return r.reconcileResources(ctx, cloudService)
}

func (r *CloudServiceReconciler) reconcileResources(ctx context.Context, cloudService *cloudServicev1alpha1.CloudService) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	manifests, err := utils.GetCloudServiceManifest(r.ComponentDir, cloudService)
	if err != nil {
		log.Error(err, "Failed to get manifests")
		return ctrl.Result{}, err
	}
	for _, item := range manifests.Items {
		namespacedName := k8sType.NamespacedName{Namespace: item.GetNamespace(), Name: item.GetName()}
		r.Get(ctx, namespacedName, nil)
		if _, err := getResource(ctx, r.DynamicInterface, &item); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Creating resource", "kind", item.GetKind(), "name", item.GetName())
				err = createResource(ctx, r.DynamicInterface, &item)
				if err != nil {
					log.Error(err, "Failed to create resource")
					return ctrl.Result{}, err
				}
			} else {
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Updating resource", "kind", item.GetKind(), "name", item.GetName())
			err = updateResource(ctx, r.DynamicInterface, &item)
			if err != nil {
				log.Error(err, "Failed to update resource")
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *CloudServiceReconciler) handleDeletion(ctx context.Context, cloudService *cloudServicev1alpha1.CloudService) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	if slices.Contains(cloudService.Finalizers, FinalizerName) {

		manifests, err := utils.GetCloudServiceManifest(r.ComponentDir, cloudService)
		if err != nil {
			log.Error(err, "Failed to get manifests")
			return ctrl.Result{}, err
		}
		for _, item := range manifests.Items {
			if _, err := getResource(ctx, r.DynamicInterface, &item); err == nil {
				log.Info("Deleting resource", "kind", item.GetKind(), "name", item.GetName())
				if err := deleteResource(ctx, r.DynamicInterface, &item); err != nil {
					log.Error(err, "Failed to delete resource")
					return ctrl.Result{}, err
				}
			}
		}

		// Remove Finalizer
		cloudService.Finalizers = removeString(cloudService.Finalizers, FinalizerName)
		if err := r.Update(ctx, cloudService); err != nil {
			log.Error(err, "Failed to remove Finalizer")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudServicev1alpha1.CloudService{}).
		Named("cloudservice").
		Complete(r)
}
