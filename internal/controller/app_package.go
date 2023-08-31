package controller

import (
	"context"
	"fmt"
	"os"

	fe "emperror.dev/errors"
	operatoroceaniov1alpha1 "github.com/f-rambo/operatorapp/api/v1alpha1"
	utils "github.com/f-rambo/operatorapp/utils"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	RepositoryConfig = "repository/repositories.yaml"
	RepositoryCache  = "repository/cache"
	HelmStorage      = "configmap"
)

func (r *AppReconciler) getCliSetting() *cli.EnvSettings {
	settings := cli.New()
	settings.KubeConfig = clientcmd.RecommendedHomeFile
	settings.RepositoryConfig = RepositoryConfig
	settings.RepositoryCache = RepositoryCache
	settings.Debug = true
	settings.KubeAPIServer = r.Cfg.Host
	settings.KubeToken = r.Cfg.BearerToken
	settings.KubeCaFile = r.Cfg.TLSClientConfig.CAFile
	return settings
}

func (r *AppReconciler) getRepoEntry(ctx context.Context, app *operatoroceaniov1alpha1.App, secret *corev1.Secret) (*repo.Entry, error) {
	logger := r.Log
	repoEntry := &repo.Entry{
		Name: app.Spec.RepoName,
		URL:  app.Spec.RepoURL,
	}
	if secret != nil && secret.Data != nil {
		for k, v := range secret.Data {
			logger.Info("secret", "key", k, "value", string(v))
			switch k {
			case "username":
				repoEntry.Username = string(v)
			case "password":
				repoEntry.Password = string(v)
			case "certFile":
				repoEntry.CertFile = string(v)
			case "keyFile":
				repoEntry.KeyFile = string(v)
			case "caFile":
				repoEntry.CAFile = string(v)
			case "insecureSkipTLSverify":
				repoEntry.InsecureSkipTLSverify = string(v) == "true"
			case "passCredentialsAll":
				repoEntry.PassCredentialsAll = string(v) == "true"
			}
		}
	}
	return repoEntry, nil
}

func (r *AppReconciler) fatchRepo(ctx context.Context, app *operatoroceaniov1alpha1.App, secret *corev1.Secret) error {
	logger := r.Log
	repoEntry, err := r.getRepoEntry(ctx, app, secret)
	if err != nil {
		return err
	}
	settings := r.getCliSetting()
	if !utils.CheckFileIsExist(settings.RepositoryCache) {
		err = utils.CreateDir(settings.RepositoryCache)
		if err != nil {
			return err
		}
	}
	if !utils.CheckFileIsExist(settings.RepositoryConfig) {
		err = utils.CreateFile(settings.RepositoryConfig)
		if err != nil {
			return err
		}
	}
	f, err := repo.LoadFile(settings.RepositoryConfig)
	if err != nil {
		return err
	}
	chart, err := repo.NewChartRepository(repoEntry, getter.All(settings))
	if err != nil {
		return err
	}
	chart.CachePath = settings.RepositoryCache
	fname, err := chart.DownloadIndexFile()
	if err != nil {
		return err
	}
	logger.Info("Create the index file in the cache directory :" + fname)
	if os.IsNotExist(fe.Cause(err)) || len(f.Repositories) == 0 {
		f = repo.NewFile()
	}
	if f == nil {
		return fmt.Errorf("repository file is empty")
	}
	f.Update(repoEntry)
	return f.WriteFile(settings.RepositoryConfig, 0644)
}

func (r *AppReconciler) deployApp(ctx context.Context, app *operatoroceaniov1alpha1.App, configMap *corev1.ConfigMap) error {
	// logger := log.FromContext(ctx)
	appConfigValues := make(map[string]interface{})
	if val, ok := configMap.Data["config"]; ok {
		err := yaml.Unmarshal([]byte(val), &appConfigValues)
		if err != nil {
			return err
		}
	}
	settings := r.getCliSetting()
	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), HelmStorage, r.Log.Infof)
	if err != nil {
		return err
	}
	chartOption := &action.ChartPathOptions{
		Version: app.Spec.Version,
	}
	chartPath, err := chartOption.LocateChart(app.Spec.ChartName, settings)
	if err != nil {
		return err
	}
	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	// list
	list := action.NewList(actionConfig)
	list.Deployed = true
	resList, err := list.Run()
	if err != nil {
		return err
	}
	isInstalled := false
	for _, v := range resList {
		if v.Name == app.Spec.ReleaseName {
			isInstalled = true
			break
		}
	}
	// install or upgrade
	if isInstalled {
		upgrade := action.NewUpgrade(actionConfig)
		upgrade.Atomic = true
		upgrade.Install = true
		upgrade.Namespace = app.Namespace
		_, err = upgrade.Run(app.Spec.ReleaseName, chart, appConfigValues)
		return err
	}
	install := action.NewInstall(actionConfig)
	install.ReleaseName = app.Spec.ReleaseName
	install.Namespace = app.Namespace
	install.CreateNamespace = true
	_, err = install.Run(chart, appConfigValues)
	if err != nil {
		return err
	}
	return nil
}

func (r *AppReconciler) deleteApp(ctx context.Context, app *operatoroceaniov1alpha1.App) error {
	actionConfig := new(action.Configuration)
	settings := r.getCliSetting()
	err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), HelmStorage, r.Log.Infof)
	if err != nil {
		return err
	}
	// check if the release is installed
	list := action.NewList(actionConfig)
	list.Deployed = true
	resList, err := list.Run()
	if err != nil {
		return err
	}
	isInstalled := false
	for _, v := range resList {
		if v.Name == app.Spec.ReleaseName {
			isInstalled = true
			break
		}
	}
	if !isInstalled {
		return nil
	}
	uninstall := action.NewUninstall(actionConfig)
	uninstall.KeepHistory = true
	uninstall.Wait = true
	_, err = uninstall.Run(app.Spec.ReleaseName)
	return err
}
