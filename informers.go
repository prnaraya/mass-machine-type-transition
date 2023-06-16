package mass_machine_type_transition

import (
	"fmt"
	"time"
	
	"k8s.io/apimachinery/pkg/fields"
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
	
	//check if deleted VMI is in list of VMIs that need to be restarted
	_, exists := vmisPendingUpdate[vmiKey]
	if  !exists {
		return
	}
	
	// remove deleted VMI from list
	delete(vmisPendingUpdate, vmiKey)
	
	// check if VMI list is now empty, to signal exiting the job
	if len(vmisPendingUpdate) == 0 {
		close(exitJob)
	}
}
