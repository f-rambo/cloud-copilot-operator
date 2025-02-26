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

	cloudcopilotv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudServiceReconciler reconciles a CloudService object
type CloudServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	cloudService := &cloudcopilotv1alpha1.CloudService{}
	err := r.Get(ctx, req.NamespacedName, cloudService)
	if err != nil {
		if errors.IsNotFound(err) {
			// CR no longer exists, remove finalizer if it exists
			log.Info("CloudService deleted, cleaning up resources", "namespace", req.Namespace, "name", req.Name)
			return r.cleanupResources(ctx, req)
		}
		log.Error(err, "Failed to get CloudService")
		return ctrl.Result{}, err
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
		return r.handleDeletion(ctx, cloudService, req)
	}

	// Create or update resources
	return r.reconcileResources(ctx, cloudService, req)
}

func (r *CloudServiceReconciler) cleanupResources(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Deployment
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err == nil {
		log.Info("Deleting Deployment", "name", deployment.Name)
		if err := r.Delete(ctx, deployment); err != nil {
			log.Error(err, "Failed to delete Deployment")
			return ctrl.Result{}, err
		}
	}

	// StatefulSet
	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(ctx, req.NamespacedName, statefulSet); err == nil {
		log.Info("Deleting StatefulSet", "name", statefulSet.Name)
		if err := r.Delete(ctx, statefulSet); err != nil {
			log.Error(err, "Failed to delete StatefulSet")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *CloudServiceReconciler) handleDeletion(ctx context.Context, cloudService *cloudcopilotv1alpha1.CloudService, _ ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	if slices.Contains(cloudService.Finalizers, FinalizerName) {
		if len(cloudService.Spec.Volumes) == 0 {
			deployment := NewDeployment(cloudService)
			if err := r.Delete(ctx, deployment); err != nil && !errors.IsNotFound(err) {
				log.Error(err, "Failed to delete Deployment")
				return ctrl.Result{}, err
			}
		}
		if len(cloudService.Spec.Volumes) != 0 {
			statefulSet := NewStatefulSet(cloudService)
			if err := r.Delete(ctx, statefulSet); err != nil && !errors.IsNotFound(err) {
				log.Error(err, "Failed to delete StatefulSet")
				return ctrl.Result{}, err
			}
		}

		configMap := NewConfigMap(cloudService)
		if err := r.Delete(ctx, configMap); err != nil && !errors.IsNotFound(err) {
			log.Error(err, "Failed to delete ConfigMap")
			return ctrl.Result{}, err
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

func (r *CloudServiceReconciler) reconcileResources(ctx context.Context, cloudService *cloudcopilotv1alpha1.CloudService, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if len(cloudService.Spec.Volumes) == 0 {
		deployment := NewDeployment(cloudService)
		if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Creating Deployment", "name", deployment.Name)
				err = r.Create(ctx, deployment)
				if err != nil {
					log.Error(err, "Failed to create Deployment")
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		}
		log.Info("Updating Deployment", "name", deployment.Name)
		err := r.Update(ctx, deployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment")
			return ctrl.Result{}, err
		}
	}

	if len(cloudService.Spec.Volumes) != 0 {
		statefulSet := NewStatefulSet(cloudService)
		if err := r.Get(ctx, req.NamespacedName, statefulSet); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Creating StatefulSet", "name", statefulSet.Name)
				err = r.Create(ctx, statefulSet)
				if err != nil {
					log.Error(err, "Failed to create StatefulSet")
					return ctrl.Result{}, err
				}
			}
			return ctrl.Result{}, nil
		}
		log.Info("Updating StatefulSet", "name", statefulSet.Name)
		err := r.Update(ctx, statefulSet)
		if err != nil {
			log.Error(err, "Failed to update StatefulSet")
			return ctrl.Result{}, err
		}
	}

	configMap := NewConfigMap(cloudService)
	if err := r.Get(ctx, req.NamespacedName, configMap); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ConfigMap", "name", configMap.Name)
			if err := r.Create(ctx, configMap); err != nil {
				log.Error(err, "Failed to create ConfigMap")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	} else {
		log.Info("Updating ConfigMap", "name", configMap.Name)
		if err := r.Update(ctx, configMap); err != nil {
			log.Error(err, "Failed to update ConfigMap")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func removeString(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcopilotv1alpha1.CloudService{}).
		Named("cloudservice").
		Complete(r)
}
