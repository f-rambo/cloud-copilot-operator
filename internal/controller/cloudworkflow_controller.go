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

	argoworkflow "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow"
	cloudcopilotv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
	"github.com/f-rambo/cloud-copilot/operator/utils"
	coreV1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	mateV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudWorkflowReconciler reconciles a CloudWorkflow object
type CloudWorkflowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudworkflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudworkflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudworkflows/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CloudWorkflow object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *CloudWorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// TODO(user): your logic here
	cloudWorkflow := &cloudcopilotv1alpha1.CloudWorkflow{}
	err := r.Get(ctx, req.NamespacedName, cloudWorkflow)
	if err != nil {
		if k8sError.IsNotFound(err) {
			err = CleanWorkflow(ctx, req.NamespacedName.Name, req.NamespacedName.Namespace)
			if err != nil {
				log.Error(err, "clean workflow failed")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if cloudWorkflow.Status.Status == cloudcopilotv1alpha1.WorkflowStatusPending {
		err = CommitWorkflow(ctx, &cloudWorkflow.Spec)
		if err != nil {
			log.Error(err, "commit workflow failed")
			return ctrl.Result{}, err
		}
	}
	if cloudWorkflow.Status.Status == cloudcopilotv1alpha1.WorkflowStatusRunning {
		for _, task := range cloudWorkflow.Spec.GetAllTasksByOrder() {
			if task.Status == cloudcopilotv1alpha1.WorkfloStatus_UNSPECIFIED {
				task.Status = cloudcopilotv1alpha1.WorkfloStatus_Pending
			}
		}
		err = r.Update(ctx, cloudWorkflow)
		if err != nil {
			log.Error(err, "update workflow status failed")
			return ctrl.Result{}, err
		}

		// Set and wait task status
		err = SetWorkflow(ctx, &cloudWorkflow.Spec)
		if err != nil {
			log.Error(err, "set workflow status failed")
			return ctrl.Result{}, err
		}
		taskPendingNumber := 0
		for _, step := range cloudWorkflow.Spec.WorkflowSteps {
			for _, task := range step.WorkflowTasks {
				if task.Status == cloudcopilotv1alpha1.WorkfloStatus_Failure {
					cloudWorkflow.Status.Status = cloudcopilotv1alpha1.WorkflowStatusFaile
					break
				}
				if task.Status == cloudcopilotv1alpha1.WorkfloStatus_Pending {
					taskPendingNumber++
				}
			}
		}
		if cloudWorkflow.Status.Status != cloudcopilotv1alpha1.WorkflowStatusFaile && taskPendingNumber == 0 {
			cloudWorkflow.Status.Status = cloudcopilotv1alpha1.WorkflowStatusSucceede
		}
		err = r.Update(ctx, cloudWorkflow)
		if err != nil {
			log.Error(err, "update workflow status failed")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func CommitWorkflow(ctx context.Context, wf *cloudcopilotv1alpha1.Workflow) error {
	argoClient, err := utils.NewArgoWorkflowClient(wf.Namespace)
	if err != nil {
		return err
	}
	argoWf, err := argoClient.Create(ctx, utils.ConvertToArgoWorkflow(wf))
	if err != nil {
		return err
	}
	kubeClient, err := utils.GetKubeClientByKubeConfig()
	if err != nil {
		return err
	}
	pvc := &coreV1.PersistentVolumeClaim{
		ObjectMeta: mateV1.ObjectMeta{
			Name:      wf.GetStorageName(),
			Namespace: wf.Namespace,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			AccessModes: []coreV1.PersistentVolumeAccessMode{
				coreV1.ReadWriteOnce,
			},
			Resources: coreV1.VolumeResourceRequirements{
				Requests: coreV1.ResourceList{
					coreV1.ResourceStorage: *resource.NewQuantity(int64(1)*GI, resource.BinarySI),
				},
			},
			StorageClassName: &wf.StorageClass,
		},
	}
	pvc.OwnerReferences = []mateV1.OwnerReference{
		{
			APIVersion:         argoworkflow.APIVersion,
			Kind:               argoworkflow.WorkflowKind,
			Name:               argoWf.Name,
			UID:                argoWf.UID,
			Controller:         utils.PointerBool(true),
			BlockOwnerDeletion: utils.PointerBool(true),
		},
	}
	_, err = kubeClient.CoreV1().PersistentVolumeClaims(wf.Namespace).Create(ctx, pvc, mateV1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func SetWorkflow(ctx context.Context, wf *cloudcopilotv1alpha1.Workflow) error {
	argoClient, err := utils.NewArgoWorkflowClient(wf.Namespace)
	if err != nil {
		return err
	}
	tasks := wf.GetAllTasksByOrder()
	for _, task := range tasks {
		if task.Status == cloudcopilotv1alpha1.WorkfloStatus_Pending {
			success, waitErr := argoClient.WaitNodeStatusByTaskName(ctx, wf.Name, task.Name)
			if waitErr != nil {
				return waitErr
			}
			if success {
				task.Status = cloudcopilotv1alpha1.WorkfloStatus_Success
			} else {
				task.Status = cloudcopilotv1alpha1.WorkfloStatus_Failure
			}
		}
	}
	argoWf, err := argoClient.Get(ctx, wf.Name, mateV1.GetOptions{})
	if err != nil {
		return err
	}
	utils.SetWorkflowStatus(argoWf, wf)
	return nil
}

func CleanWorkflow(ctx context.Context, name, namespace string) error {
	wf := &cloudcopilotv1alpha1.Workflow{
		Name:      name,
		Namespace: namespace,
	}
	argoClient, err := utils.NewArgoWorkflowClient(wf.Namespace)
	if err != nil {
		return err
	}
	err = argoClient.Delete(ctx, wf.Name)
	if err != nil && !k8sError.IsNotFound(err) {
		return err
	}
	kubeClient, err := utils.GetKubeClientByKubeConfig()
	if err != nil {
		return err
	}
	err = kubeClient.CoreV1().PersistentVolumeClaims(wf.Namespace).Delete(ctx, wf.GetStorageName(), mateV1.DeleteOptions{})
	if err != nil && !k8sError.IsNotFound(err) {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudWorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcopilotv1alpha1.CloudWorkflow{}).
		Named("cloudworkflow").
		Complete(r)
}
