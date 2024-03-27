package controller

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	operatoroceaniov1alpha1 "github.com/f-rambo/operatorapp/api/v1alpha1"
	"github.com/f-rambo/operatorapp/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *AppReconciler) deployment(ctx context.Context, req ctrl.Request, app *operatoroceaniov1alpha1.App) error {
	deployment, err := utils.NewDeployment(app)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(app, deployment, r.Scheme); err != nil {
		return err
	}
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
	return nil
}
