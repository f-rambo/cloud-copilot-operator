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

func (r *AppReconciler) deployServiceApp(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) (err error) {
	fs := []func(context.Context, ctrl.Request, *operatoroceaniov1alpha1.App) error{
		r.configmap,
		r.secret,
		r.deployment,
		r.service,
		r.ingress,
	}
	for _, f := range fs {
		err = f(ctx, req, app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *AppReconciler) configmap(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) error {
	configMap := utils.NewConfigMap(app)
	err := controllerutil.SetControllerReference(app, configMap, r.Scheme)
	if err != nil {
		return err
	}
	c := &corev1.ConfigMap{}
	err = r.Get(ctx, req.NamespacedName, c)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		if err := r.Create(ctx, configMap); err != nil {
			return err
		}
		return nil
	}
	return r.Update(ctx, configMap)
}

func (r *AppReconciler) secret(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) error {
	secret := utils.NewSecret(app)
	err := controllerutil.SetControllerReference(app, secret, r.Scheme)
	if err != nil {
		return err
	}
	s := &corev1.Secret{}
	err = r.Get(ctx, req.NamespacedName, s)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		if err := r.Create(ctx, secret); err != nil {
			return err
		}
		return nil
	}
	return r.Update(ctx, secret)
}

func (r *AppReconciler) deployment(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) error {
	deployment, err := utils.NewDeployment(app)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(app, deployment, r.Scheme); err != nil {
		return err
	}
	//查找同名deployment
	d := &appv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, d); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, deployment); err != nil {
				return err
			}
		}
	} else {
		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
	}
	return nil
}

func (r *AppReconciler) service(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) error {
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
				return err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.Service.EnableService {
			return err
		}
	} else {
		if app.Spec.Service.EnableService {
			if err := r.Update(ctx, service); err != nil {
				return err
			}
		} else {
			if err := r.Delete(ctx, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *AppReconciler) ingress(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) error {
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
				return err
			}
		}
		if !errors.IsNotFound(err) && app.Spec.Service.EnableIngress {
			return err
		}
	} else {
		if app.Spec.Service.EnableIngress {
			if err := r.Update(ctx, ingress); err != nil {
				return err
			}
		} else {
			if err := r.Delete(ctx, i); err != nil {
				return err
			}
		}
	}
	return nil
}
