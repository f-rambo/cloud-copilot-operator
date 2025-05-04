package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	pkgChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	helmValues "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var settings = cli.New()

type HelmPkg struct {
	actionConfig *action.Configuration
	logs         []string
}

func NewHelmPkg(namespace string) (*HelmPkg, error) {
	settings.Debug = true
	if namespace == "" {
		namespace = "default"
	}
	helmPkg := &HelmPkg{logs: make([]string, 0)}
	settings.SetNamespace(namespace)
	kube.ManagedFieldsManager = "helm"
	actionConfig := new(action.Configuration)
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), helmDriver, helmPkg.Logf); err != nil {
		return nil, err
	}
	if helmDriver == "memory" {
		loadReleasesInMemory(actionConfig)
	}
	registryClient, err := newDefaultRegistryClient(false)
	if err != nil {
		return nil, err
	}
	actionConfig.RegistryClient = registryClient
	helmPkg.actionConfig = actionConfig
	return helmPkg, nil
}

func (d *HelmPkg) Logf(format string, v ...interface{}) {
	log := fmt.Sprintf(format, v...)

	if !slices.Contains(d.logs, log) {
		d.logs = append(d.logs, log)
	}
}

func (h *HelmPkg) GetLogs() string {
	return strings.Join(h.logs, "\n")
}

func newDefaultRegistryClient(plainHTTP bool) (*registry.Client, error) {
	opts := []registry.ClientOption{
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptEnableCache(true),
		registry.ClientOptWriter(os.Stderr),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	}
	if plainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	// Create a new registry client
	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newRegistryClient(certFile, keyFile, caFile string, insecureSkipTLSverify, plainHTTP bool) (*registry.Client, error) {
	if certFile != "" && keyFile != "" || caFile != "" || insecureSkipTLSverify {
		registryClient, err := newRegistryClientWithTLS(certFile, keyFile, caFile, insecureSkipTLSverify)
		if err != nil {
			return nil, err
		}
		return registryClient, nil
	}
	registryClient, err := newDefaultRegistryClient(plainHTTP)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

func newRegistryClientWithTLS(certFile, keyFile, caFile string, insecureSkipTLSverify bool) (*registry.Client, error) {
	// Create a new registry client
	registryClient, err := registry.NewRegistryClientWithTLS(os.Stderr, certFile, keyFile, caFile, insecureSkipTLSverify,
		settings.RegistryConfig, settings.Debug,
	)
	if err != nil {
		return nil, err
	}
	return registryClient, nil
}

// This function loads releases into the memory storage if the
// environment variable is properly set.
func loadReleasesInMemory(actionConfig *action.Configuration) {
	filePaths := strings.Split(os.Getenv("HELM_MEMORY_DRIVER_DATA"), ":")
	if len(filePaths) == 0 {
		return
	}

	store := actionConfig.Releases
	mem, ok := store.Driver.(*driver.Memory)
	if !ok {
		// For an unexpected reason we are not dealing with the memory storage driver.
		return
	}

	actionConfig.KubeClient = &kubefake.PrintingKubeClient{Out: io.Discard}

	for _, path := range filePaths {
		b, err := os.ReadFile(path)
		if err != nil {
			log.Fatal("Unable to read memory driver data", err)
		}

		releases := []*release.Release{}
		if err := yaml.Unmarshal(b, &releases); err != nil {
			log.Fatal("Unable to unmarshal memory driver data: ", err)
		}

		for _, rel := range releases {
			if err := store.Create(rel); err != nil {
				log.Fatal(err)
			}
		}
	}
	// Must reset namespace to the proper one
	mem.SetNamespace(settings.Namespace())
}

func (h *HelmPkg) Write(p []byte) (n int, err error) {
	h.Logf("%s", string(p))
	return len(p), nil
}

func validateDryRunOptionFlag(dryRunOptionFlagValue string) error {
	// Validate dry-run flag value with a set of allowed value
	allowedDryRunValues := []string{"false", "true", "none", "client", "server"}
	isAllowed := false
	for _, v := range allowedDryRunValues {
		if dryRunOptionFlagValue == v {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return errors.New("Invalid dry-run flag. Flag must one of the following: false, true, none, client, server")
	}
	return nil
}

// Application chart type is only installable
func checkIfInstallable(ch *pkgChart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

type ChartInfo struct {
	Name        string
	Config      string
	Readme      string
	Description string
	Metadata    pkgChart.Metadata
	Version     string
	AppName     string
	Chart       string
}

func GetLocalChartInfo(chart string) (*ChartInfo, error) {
	readme, err := action.NewShow(action.ShowReadme).Run(chart)
	if err != nil {
		return nil, errors.WithMessage(err, "show readme fail")
	}
	chartYaml, err := action.NewShow(action.ShowChart).Run(chart)
	if err != nil {
		return nil, errors.WithMessage(err, "show chart fail")
	}
	valuesYaml, err := action.NewShow(action.ShowValues).Run(chart)
	if err != nil {
		return nil, errors.WithMessage(err, "show values fail")
	}
	chartMateData := &pkgChart.Metadata{}
	err = yaml.Unmarshal([]byte(chartYaml), chartMateData)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal chart yaml fail")
	}
	chartInfo := &ChartInfo{
		Name:        chartMateData.Name,
		Config:      valuesYaml,
		Readme:      readme,
		Description: chartMateData.Description,
		Metadata:    *chartMateData,
		Version:     chartMateData.Version,
		AppName:     chartMateData.Name,
		Chart:       chart,
	}
	return chartInfo, nil
}

func (h *HelmPkg) NewInstall() (*action.Install, error) {
	client := action.NewInstall(h.actionConfig)
	registryClient, err := newRegistryClient(client.CertFile, client.KeyFile, client.CaFile,
		client.InsecureSkipTLSverify, client.PlainHTTP)
	if err != nil {
		return nil, err
	}
	client.SetRegistryClient(registryClient)
	return client, nil
}

func (h *HelmPkg) checkIfInstall(client *action.Install) (bool, error) {
	var isExist bool = false
	list, err := h.NewList()
	if err != nil {
		return isExist, err
	}
	data, err := h.RunList(list)
	if err != nil {
		return isExist, err
	}
	for _, r := range data {
		if r.Name == client.ReleaseName {
			isExist = true
		}
	}
	return isExist, nil
}

func (h *HelmPkg) RunInstall(ctx context.Context, client *action.Install, chart string, valueOptions *helmValues.Options) (*release.Release, error) {
	ok, err := h.checkIfInstall(client)
	if err != nil {
		return nil, err
	}
	if ok {
		upgrade, getErr := h.NewUpGrade()
		if getErr != nil {
			return nil, getErr
		}
		upgrade.ReuseValues = true
		upgrade.Version = client.Version
		upgrade.ChartPathOptions = client.ChartPathOptions
		upgrade.Force = client.Force
		upgrade.DryRun = client.DryRun
		upgrade.DryRunOption = client.DryRunOption
		upgrade.DisableHooks = client.DisableHooks
		upgrade.SkipCRDs = client.SkipCRDs
		upgrade.Timeout = client.Timeout
		upgrade.Wait = client.Wait
		upgrade.WaitForJobs = client.WaitForJobs
		upgrade.Devel = client.Devel
		upgrade.Namespace = client.Namespace
		upgrade.Atomic = client.Atomic
		upgrade.PostRenderer = client.PostRenderer
		upgrade.DisableOpenAPIValidation = client.DisableOpenAPIValidation
		upgrade.SubNotes = client.SubNotes
		upgrade.Description = client.Description
		upgrade.DependencyUpdate = client.DependencyUpdate
		upgrade.Labels = client.Labels
		upgrade.EnableDNS = client.EnableDNS
		return h.RunUpgrade(ctx, upgrade, client.ReleaseName, chart, valueOptions)
	}
	if client.Timeout == 0 {
		client.Timeout = 300 * time.Second
	}
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	if client.DryRunOption == "" {
		client.DryRunOption = "none"
	}
	p := getter.All(settings)
	chartPath, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}
	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}
	if getErr := checkIfInstallable(chartRequested); getErr != nil {
		return nil, getErr
	}
	if chartRequested.Metadata.Deprecated {
		h.Logf("This chart is deprecated")
	}
	if req := chartRequested.Metadata.Dependencies; req != nil {
		if getErr := action.CheckDependencies(chartRequested, req); getErr != nil {
			err = errors.Wrap(getErr, "An error occurred while checking for chart dependencies. You may need to run `helm dependency build` to fetch missing dependencies")
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              h,
					ChartPath:        chartPath,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
					RegistryClient:   client.GetRegistryClient(),
				}
				if getErr := man.Update(); getErr != nil {
					return nil, getErr
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(chartPath); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}
	// Validate DryRunOption member is one of the allowed values
	if getErr := validateDryRunOptionFlag(client.DryRunOption); err != nil {
		return nil, getErr
	}
	values, err := valueOptions.MergeValues(p)
	if err != nil {
		return nil, err
	}
	return client.RunWithContext(ctx, chartRequested, values)
}

func (h *HelmPkg) NewList() (*action.List, error) {
	client := action.NewList(h.actionConfig)
	if client.AllNamespaces {
		if err := h.actionConfig.Init(settings.RESTClientGetter(), "", os.Getenv("HELM_DRIVER"), h.Logf); err != nil {
			return nil, err
		}
	}
	client.SetStateMask()
	return client, nil
}

func (h *HelmPkg) RunList(client *action.List) ([]*release.Release, error) {
	results, err := client.Run()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (h *HelmPkg) NewPull() (*action.Pull, error) {
	pullObj := action.NewPullWithOpts(action.WithConfig(h.actionConfig))
	pullObj.SetRegistryClient(h.actionConfig.RegistryClient)
	return pullObj, nil
}

func (h *HelmPkg) RunPull(client *action.Pull, dir, chartRef, version string) error {
	client.DestDir = dir
	client.Version = version
	_, err := client.Run(chartRef)
	if err != nil {
		return err
	}
	return nil
}

func (h *HelmPkg) NewUninstall() (*action.Uninstall, error) {
	client := action.NewUninstall(h.actionConfig)
	return client, nil
}

func (h *HelmPkg) RunUninstall(client *action.Uninstall, name string) (*release.UninstallReleaseResponse, error) {
	list, err := h.NewList()
	if err != nil {
		return nil, err
	}
	data, err := h.RunList(list)
	if err != nil {
		return nil, err
	}
	var isExist bool = false
	for _, r := range data {
		if r.Name == name {
			isExist = true
		}
	}
	if !isExist {
		return &release.UninstallReleaseResponse{
			Release: &release.Release{
				Info: &release.Info{
					Status: release.StatusUninstalled,
				},
			},
		}, nil
	}
	if client.Timeout == 0 {
		client.Timeout = 300 * time.Second
	}
	return client.Run(name)
}

func (h *HelmPkg) NewUpGrade() (*action.Upgrade, error) {
	client := action.NewUpgrade(h.actionConfig)
	client.Namespace = settings.Namespace()
	registryClient, err := newRegistryClient(client.CertFile, client.KeyFile, client.CaFile,
		client.InsecureSkipTLSverify, client.PlainHTTP)
	if err != nil {
		return nil, err
	}
	client.SetRegistryClient(registryClient)
	return client, nil
}

func (h *HelmPkg) RunUpgrade(ctx context.Context, client *action.Upgrade, name, chart string, valueOptions *helmValues.Options) (*release.Release, error) {
	if client.Timeout == 0 {
		client.Timeout = 300 * time.Second
	}
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	if client.DryRunOption == "" {
		client.DryRunOption = "none"
	}
	chartPath, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return nil, err
	}
	// Validate dry-run flag value is one of the allowed values
	if getErr := validateDryRunOptionFlag(client.DryRunOption); getErr != nil {
		return nil, getErr
	}

	p := getter.All(settings)

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if getErr := action.CheckDependencies(ch, req); getErr != nil {
			err = errors.Wrap(getErr, "An error occurred while checking for chart dependencies. You may need to run `helm dependency build` to fetch missing dependencies")
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              h,
					ChartPath:        chartPath,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if getErr := man.Update(); getErr != nil {
					return nil, getErr
				}
				// Reload the chart with the updated Chart.lock file.
				if ch, err = loader.Load(chartPath); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}

	if ch.Metadata.Deprecated {
		h.Logf("This chart is deprecated")
	}
	values, err := valueOptions.MergeValues(p)
	if err != nil {
		return nil, err
	}
	return client.RunWithContext(ctx, name, ch, values)
}

func (h *HelmPkg) GetReleaseInfo(releaseName string) (*release.Release, error) {
	get := action.NewGet(h.actionConfig)
	return get.Run(releaseName)
}
