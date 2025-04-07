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

	cloudcopilotv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudWorkspaceReconciler reconciles a CloudWorkspace object
type CloudWorkspaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudworkspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudworkspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudworkspaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CloudWorkspace object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *CloudWorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	cloudWorkspace := &cloudcopilotv1alpha1.CloudWorkspace{}
	if err := r.Get(ctx, req.NamespacedName, cloudWorkspace); err != nil {
		if k8sErr.IsNotFound(err) {
			err = r.DeleteNamespace(ctx, cloudWorkspace)
			if err != nil {
				log.Error(err, "failed to delete namespace")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info(cloudWorkspace.Name)
	err := r.CreateNamespaceOrUpdateNamespace(ctx, cloudWorkspace)
	if err != nil {
		log.Error(err, "failed to create or update namespace")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// CreateNamespaceOrUpdateNamespace
// CreateNamespaceOrUpdateNamespace creates or updates a namespace for the CloudWorkspace
func (r *CloudWorkspaceReconciler) CreateNamespaceOrUpdateNamespace(ctx context.Context, cloudWorkspace *cloudcopilotv1alpha1.CloudWorkspace) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cloudWorkspace.Name,
			Labels: map[string]string{
				"cloudworkspace": cloudWorkspace.Name,
				"created-by":     "cloud-copilot-operator",
			},
		},
	}

	err := r.Create(ctx, namespace)
	if err != nil {
		if k8sErr.IsAlreadyExists(err) {
			existing := &corev1.Namespace{}
			if getErr := r.Get(ctx, client.ObjectKey{Name: cloudWorkspace.Name}, existing); getErr != nil {
				return getErr
			}

			existing.Labels = namespace.Labels
			return r.Update(ctx, existing)
		}
		return err
	}
	return nil
}

// DeleteNamespace
// DeleteNamespace deletes the namespace associated with the CloudWorkspace
func (r *CloudWorkspaceReconciler) DeleteNamespace(ctx context.Context, cloudWorkspace *cloudcopilotv1alpha1.CloudWorkspace) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: cloudWorkspace.Name,
		},
	}

	err := r.Delete(ctx, namespace)
	if err != nil && !k8sErr.IsNotFound(err) {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudWorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcopilotv1alpha1.CloudWorkspace{}).
		Named("cloudworkspace").
		Complete(r)
}
