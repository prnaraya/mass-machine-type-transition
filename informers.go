package mass_machine_type_transition

import (
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
	listWatcher := cache.NewListWatchFromClient(virtCli.RestClient(), "virtualmachineinstances", namespace, fields.Everything())
	vmiInformer := cache.NewSharedIndexInformer(listWatcher, &k6tv1.VirtualMachineInstance{}, 1*time.Hour, cache.Indexers{})
	
	vmiInformer.AddEventHandler(cache.ResourceEventHandlerFuncs {
		DeleteFunc: handleDeletedVmi,
	})

	return vmiInformer, nil
}

func handleDeletedVmi(obj interface{}) {
	vmi, ok := obj.(*k6tv1.VirtualMachineInstance)
	if !ok {
		return
	}
	
	// get VMI name in the format namespace/name
	vmiKey, err := cache.MetaNamespaceKeyFunc(vmi)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	// check if deleted VMI is in list of VMIs that need to be restarted
	_, exists := vmisPendingUpdate[vmiKey]
	if  !exists {
		return
	}
	
	// remove warning label from VM
	// if removing the warning label fails, exit before removing VMI from list
	// since the label is still there to tell the user to restart, it wouldn't
	// make sense to have a mismatch between the number of VMs with the label
	// and the number of VMIs in the list of VMIs pending update.
	
	err = removeWarningLabel(vmi)
	if err != nil {
		fmt.Println(err)
		
		// unit tests produce return a kubeconfig error, this ignores the error
		// to allow unit test to check for VMI removal in general
		if !ignoreKubeClientError {
			return
		}
	}
	
	// remove deleted VMI from list
	delete(vmisPendingUpdate, vmiKey)
	
	// check if VMI list is now empty, to signal exiting the job
	if len(vmisPendingUpdate) == 0 {
		close(exitJob)
	}
}

func removeWarningLabel(vmi *k6tv1.VirtualMachineInstance) error {
	virtCli, err := getVirtCli()
	if err != nil {
		return err
	}
	
	removeLabel := fmt.Sprint(`{"op": "remove", "path": "/metadata/labels/restart-vm-required"}`)
	_, err = virtCli.VirtualMachine(vmi.Namespace).Patch(vmi.Name, types.JSONPatchType, []byte(removeLabel), &k8sv1.PatchOptions{})
	return err
}
