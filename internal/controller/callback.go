package controller

import (
	"context"

	operatoroceaniov1alpha1 "github.com/f-rambo/sailor/api/v1alpha1"
	cstor "github.com/openebs/api/v3/pkg/apis/cstor/v1"
	openebsapis "github.com/openebs/api/v3/pkg/apis/openebs.io/v1alpha1"
	matev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *AppReconciler) getAppCallback(appname string) func(ctx context.Context, app *operatoroceaniov1alpha1.App) error {
	switch appname {
	case "openebs":
		return r.openebs
	}
	return nil
}

func (r *AppReconciler) openebs(ctx context.Context, app *operatoroceaniov1alpha1.App) error {
	bds := openebsapis.BlockDeviceList{}
	err := r.ClientSet.RESTClient().Get().Resource("BlockDevices").Namespace(app.Namespace).Do(ctx).Into(&bds)
	if err != nil {
		return err
	}
	if len(bds.Items) == 0 {
		return nil
	}
	cStorStoragePool := cstor.CStorPoolCluster{
		TypeMeta: matev1.TypeMeta{
			APIVersion: "cstor.openebs.io/v1",
			Kind:       "CStorPoolCluster",
		},
		ObjectMeta: matev1.ObjectMeta{
			Name:      "cstor-disk-pool",
			Namespace: app.Namespace,
		},
	}
	pools := make([]cstor.PoolSpec, 0)
	for _, bd := range bds.Items {
		poolSpec := cstor.PoolSpec{}
		poolSpec.NodeSelector = make(map[string]string)
		poolSpec.NodeSelector["kubernetes.io/hostname"] = bd.Spec.NodeAttributes.NodeName
		poolSpec.DataRaidGroups = make([]cstor.RaidGroup, 0)
		raidGroup := cstor.RaidGroup{}
		raidGroup.CStorPoolInstanceBlockDevices = make([]cstor.CStorPoolInstanceBlockDevice, 0)
		raidGroup.CStorPoolInstanceBlockDevices = append(raidGroup.CStorPoolInstanceBlockDevices, cstor.CStorPoolInstanceBlockDevice{
			BlockDeviceName: bd.Name,
		})
		poolSpec.DataRaidGroups = append(poolSpec.DataRaidGroups, raidGroup)
		pools = append(pools, poolSpec)
	}
	cStorStoragePool.Spec.Pools = pools
	err = r.ClientSet.RESTClient().Post().Resource("CStorPoolClusters").Namespace(app.Namespace).Body(&cStorStoragePool).Do(ctx).Into(&cStorStoragePool)
	if err != nil {
		return err
	}
	return nil
}
