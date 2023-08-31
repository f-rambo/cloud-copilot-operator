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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	operatoroceaniov1alpha1 "github.com/f-rambo/operatorapp/api/v1alpha1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=operator.ocean.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.ocean.io,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.ocean.io,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling App")
	// 获取app资源
	app := &operatoroceaniov1alpha1.App{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("App resource not found. Ignoring since object must be deleted")
			app.Status.Running = false
			err = r.updateAppStatus(ctx, app)
			if err != nil {
				return ctrl.Result{}, err
			}
			err = deleteApp(ctx, app)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if app.Status.Running {
		logger.Info("App already running")
		return ctrl.Result{}, nil
	}
	appConfigMap, err := r.appRefConfigMap(ctx, app)
	if err != nil {
		return ctrl.Result{}, err
	}
	appSecret, err := r.appRefSecret(ctx, app)
	if err != nil {
		return ctrl.Result{}, err
	}
	// 获取helm repo资源
	err = fatchRepo(ctx, app, appSecret)
	if err != nil {
		return ctrl.Result{}, err
	}
	// 安装helm chart
	err = deployApp(ctx, app, appConfigMap)
	if err != nil {
		return ctrl.Result{}, err
	}
	// 更新app状态
	app.Status.Running = true
	err = r.updateAppStatus(ctx, app)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatoroceaniov1alpha1.App{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func (r *AppReconciler) updateAppStatus(ctx context.Context, app *operatoroceaniov1alpha1.App) error {
	logger := log.FromContext(ctx)
	// 更新app资源状态
	err := r.Status().Update(ctx, app)
	if err != nil {
		logger.Error(err, "Failed to update App status")
		return err
	}
	return nil
}

func (r *AppReconciler) appRefConfigMap(ctx context.Context, app *operatoroceaniov1alpha1.App) (*corev1.ConfigMap, error) {
	logger := log.FromContext(ctx)
	// 获取configmap资源，并关联到app资源
	configmap := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Spec.ConfigMapName,
	}, configmap)
	if err != nil && !errors.IsNotFound(err) {
		return configmap, err
	}
	if !errors.IsNotFound(err) {
		logger.Info("ConfigMap already exists")
		err = controllerutil.SetControllerReference(app, configmap, r.Scheme)
		if err != nil {
			return configmap, err
		}
		err = r.Update(ctx, configmap)
		if err != nil {
			return configmap, err
		}
	}
	return configmap, nil
}

func (r *AppReconciler) appRefSecret(ctx context.Context, app *operatoroceaniov1alpha1.App) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)
	// 获取secret资源，并关联到app资源
	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      app.Spec.SecretName,
	}, secret)
	if err != nil && !errors.IsNotFound(err) {
		return secret, err
	}
	if !errors.IsNotFound(err) {
		logger.Info("Secret already exists")
		err = controllerutil.SetControllerReference(app, secret, r.Scheme)
		if err != nil {
			return secret, err
		}
		err = r.Update(ctx, secret)
		if err != nil {
			return secret, err
		}
	}
	return secret, nil
}
