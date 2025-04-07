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
	"encoding/json"

	"github.com/f-rambo/cloud-copilot/operator/utils"
	helmCli "helm.sh/helm/v3/pkg/cli"
	helmgetter "helm.sh/helm/v3/pkg/getter"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudcopilotv1alpha1 "github.com/f-rambo/cloud-copilot/operator/api/v1alpha1"
)

var AppPath = "apps"
var AppConfigDir = "config/apps"
var ComponentDir = "component"

// CloudAppReconciler reconciles a CloudApp object
type CloudAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud-copilot.operator.io,resources=cloudapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CloudApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *CloudAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	app := &cloudcopilotv1alpha1.CloudApp{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !app.Spec.AppIsEmpty() && app.Status.Status == cloudcopilotv1alpha1.AppStatus_PENDING {
		err = GetAppAndVersionInfo(ctx, &app.Spec.App)
		if err != nil {
			app.Status.Status = cloudcopilotv1alpha1.AppStatus_FAILED
			app.Status.Message = err.Error()
			log.Error(err, "Failed to get app info")
		} else {
			app.Status.Status = cloudcopilotv1alpha1.AppStatus_COMPLETED
		}
		err = r.Update(ctx, app)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	if !app.Spec.AppRepoIsEmpty() && app.Status.Status == cloudcopilotv1alpha1.AppStatus_PENDING {
		err = ReloadAppRepo(ctx, &app.Spec.AppRepo)
		if err != nil {
			app.Status.Status = cloudcopilotv1alpha1.AppStatus_FAILED
			app.Status.Message = err.Error()
			log.Error(err, "Failed to reload app repo")
		} else {
			app.Status.Status = cloudcopilotv1alpha1.AppStatus_COMPLETED
		}
		err = r.Update(ctx, app)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CloudAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudcopilotv1alpha1.CloudApp{}).
		Named("cloudapp").
		Complete(r)
}

func GetAppAndVersionInfo(ctx context.Context, app *cloudcopilotv1alpha1.App) error {
	for _, appVersion := range app.Versions {
		if appVersion == nil || appVersion.Chart == "" {
			continue
		}
		charInfo, err := utils.GetLocalChartInfo(appVersion.Chart)
		if err != nil {
			return err
		}
		charInfoMetadata, err := json.Marshal(charInfo.Metadata)
		if err != nil {
			return err
		}
		app.Name = charInfo.Name
		app.Readme = charInfo.Readme
		app.Description = charInfo.Description
		app.Metadata = charInfoMetadata

		appVersion.DefaultConfig = charInfo.Config
		appVersion.Version = charInfo.Version
		appVersion.Chart = charInfo.Chart
	}
	return nil
}

func ReloadAppRepo(ctx context.Context, appRepo *cloudcopilotv1alpha1.AppRepo) error {
	settings := helmCli.New()
	res, err := helmrepo.NewChartRepository(&helmrepo.Entry{
		Name: appRepo.Name,
		URL:  appRepo.Url,
	}, helmgetter.All(settings))
	if err != nil {
		return err
	}
	res.CachePath = AppPath
	indexFile, err := res.DownloadIndexFile()
	if err != nil {
		return err
	}
	appRepo.IndexPath = indexFile

	index, err := helmrepo.LoadIndexFile(appRepo.IndexPath)
	if err != nil {
		return err
	}
	apps := make([]*cloudcopilotv1alpha1.App, 0)
	for chartName, chartVersions := range index.Entries {
		app := &cloudcopilotv1alpha1.App{Name: chartName, AppRepoId: appRepo.Id, Versions: make([]*cloudcopilotv1alpha1.AppVersion, 0)}
		for _, chartMatedata := range chartVersions {
			if len(chartMatedata.URLs) == 0 {
				continue
			}
			app.Icon = chartMatedata.Icon
			app.Description = chartMatedata.Description
			appVersion := &cloudcopilotv1alpha1.AppVersion{Name: chartMatedata.Name, Chart: chartMatedata.URLs[0], Version: chartMatedata.Version}
			app.Versions = append(app.Versions, appVersion)
		}
		apps = append(apps, app)
	}
	for _, app := range apps {
		app.AppRepoId = appRepo.Id
		err = GetAppAndVersionInfo(ctx, app)
		if err != nil {
			return err
		}
	}
	appRepo.Apps = apps
	return nil
}
