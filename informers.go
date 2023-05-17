package mass_machine_type_transition

import (
	"context"
	"fmt"
	"time"
	
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	k6tv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

func getVirtCli() (kubecli.KubevirtClient, error) {
	clientConfig, err := kubecli.GetKubevirtClientConfig()
	if err != nil {
		return nil, err
	}

	virtCli, err := kubecli.GetKubevirtClientFromRESTConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return virtCli, err
}

func getVmiInformer(virtCli kubecli.KubevirtClient) (cache.SharedIndexInformer, error) {
	listWatcher := cache.NewListWatchFromClient(virtCli.RestClient(), "virtualmachineinstances", k8sv1.NamespaceAll, fields.Everything())
	vmiInformer := cache.NewSharedIndexInformer(listWatcher, &k6tv1.VirtualMachineInstance{}, 1*time.Hour, cache.Indexers{})
	
	vmiInformer.AddEventHandler(cache.ResourceEventHandlerFuncs {
		DeleteFunc: removeWarningLabel,
	})

	return vmiInformer, nil
}

// this is a temporary function that performs the mass machine type transition by manually iterating through each vm and changing the spec
func updateMachineTypes(virtCli kubecli.KubevirtClient) error {
	updateMachineType := fmt.Sprint(`{"spec":{"template":{"spec":{"domain":{"machine":{"type": "pc-q35-rhel9.0.0"}}}}}}`)
	vmList, err := virtCli.VirtualMachine(k8sv1.NamespaceAll).List(&k8sv1.ListOptions{})
	if err != nil {
		return err
	}
	for _, vm := range vmList.Items {
		_, err = virtCli.VirtualMachine(vm.Namespace).Patch(vm.Name, types.StrategicMergePatchType, []byte(updateMachineType), &k8sv1.PatchOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func removeWarningLabel(obj interface{}) {
	vmi, ok := obj.(*k6tv1.VirtualMachineInstance)
	if !ok {
		return
	}
	virtCli, err := getVirtCli()
	if err != nil {
		return
	}
	
	removeLabel := fmt.Sprint(`{"op": "remove", "path": "metadata/labels/restart-vm-required"}`)
	virtCli.CoreV1().Nodes().Patch(context.Background(), vmi.Status.NodeName, types.JSONPatchType, []byte(removeLabel), k8sv1.PatchOptions{})
}
