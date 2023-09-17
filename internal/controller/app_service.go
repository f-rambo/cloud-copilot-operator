package controller

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatoroceaniov1alpha1 "github.com/f-rambo/operatorapp/api/v1alpha1"
	"github.com/f-rambo/operatorapp/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *AppReconciler) deployServiceApp(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) error {
	// configmap
	_, err := r.appRefConfigMap(ctx, app, app.Spec.Service.ConfigMapName)
	if err != nil {
		return err
	}
	// Deployment  bind configmap
	deployment, err := utils.NewDeployment(app)
	if err != nil {
		return err
	}
	//查找同名deployment
	d := &appv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, d); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				r.Log.Error(err, "create deploy failed")
				return err
			}
		}
	} else {
		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
	}
	// service
	service, err := utils.NewService(app)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(app, service, r.Scheme); err != nil {
		return err
	}
	//查找指定service
	s := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, s); err != nil {
		if errors.IsNotFound(err) && app.Spec.Service.EnableService {
			if err := r.Create(ctx, service); err != nil {
				r.Log.Error(err, "create service failed")
				return err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.Service.EnableService {
			return err
		}
	} else {
		if app.Spec.Service.EnableService {
			r.Log.Info("skip update")
		} else {
			if err := r.Delete(ctx, s); err != nil {
				return err
			}

		}
	}
	// ingress
	ingress, err := utils.NewIngress(app)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(app, ingress, r.Scheme); err != nil {
		return err
	}
	i := &netv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, i); err != nil {
		if errors.IsNotFound(err) && app.Spec.Service.EnableIngress {
			if err := r.Create(ctx, ingress); err != nil {
				r.Log.Error(err, "create ingress failed")
				return err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.Service.EnableIngress {
			return err
		}
	} else {
		if app.Spec.Service.EnableIngress {
			r.Log.Info("skip update")
		} else {
			if err := r.Delete(ctx, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *AppReconciler) appRefConfigMap(ctx context.Context, app *operatoroceaniov1alpha1.App, configMapName string) (*corev1.ConfigMap, error) {
	if configMapName == "" {
		return nil, nil
	}
	logger := r.Log
	// 获取configmap资源，并关联到app资源
	configmap := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      configMapName,
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

func (r *AppReconciler) appRefSecret(ctx context.Context, app *operatoroceaniov1alpha1.App, secretName string) (*corev1.Secret, error) {
	logger := r.Log
	// 获取secret资源，并关联到app资源
	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: app.Namespace,
		Name:      secretName,
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
