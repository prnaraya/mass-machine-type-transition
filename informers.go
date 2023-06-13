package mass_machine_type_transition

import (
	"fmt"
	"strings"
	"time"
	
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	k6tv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

// using this as a const allows us to easily modify the program to update if a newer version is released
// we generally want to be updating the machine types to the most recent version
const latestMachineTypeVersion = "rhel9.2.0"

var ( 
	vmisPendingUpdate = make(map[string]struct{})
	exitJob = make(chan struct{})
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
		UpdateFunc: handleVmiUpdate,
		DeleteFunc: handleVmiDeletion,
	})

	return vmiInformer, nil
}

func updateMachineTypes(virtCli kubecli.KubevirtClient) error {
	vmList, err := virtCli.VirtualMachine(k8sv1.NamespaceAll).List(&k8sv1.ListOptions{})
	if err != nil {
		return err
	}
	for _, vm := range vmList.Items {
		machineType := vm.Spec.Template.Spec.Domain.Machine.Type
		machineTypeSubstrings := strings.Split(machineType, "-")
		version := machineTypeSubstrings[2]
		if len(machineTypeSubstrings) != 3 {
			return nil
		}
		
		if strings.Contains(version, "rhel") && version < "rhel9.0.0" {
			machineTypeSubstrings[2] = latestMachineTypeVersion
			machineType = strings.Join(machineTypeSubstrings, "-")
			updateMachineType := fmt.Sprintf(`{"spec": {"template": {"spec": {"domain": {"machine": {"type": "%s"}}}}}}`, machineType)
			
			_, err = virtCli.VirtualMachine(vm.Namespace).Patch(vm.Name, types.StrategicMergePatchType, []byte(updateMachineType), &k8sv1.PatchOptions{})
			if err != nil {
				return err
			}
			
			// add label to running VMs that a restart is required for change to take place
			if vm.Status.Ready {
				err = addWarningLabel(virtCli, &vm)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func addWarningLabel (virtCli kubecli.KubevirtClient, vm *k6tv1.VirtualMachine) error {
	addLabel := fmt.Sprint(`{"metadata": {"labels": {"restart-vm-required": "true"}}}}}`)
	_, err := virtCli.VirtualMachineInstance(vm.Namespace).Patch(vm.Name, types.StrategicMergePatchType, []byte(addLabel), &k8sv1.PatchOptions{})
	if err != nil {
		return err
	}
	
	// get VM name in the format namespace/name
	vmKey, err := cache.MetaNamespaceKeyFunc(vm)
	if err != nil {
		return err
	}
	vmisPendingUpdate[vmKey] = struct{}{}
	
	return nil
}

func handleVmiUpdate(oldObj interface{}, newObj interface{}) {
	vmi, ok := newObj.(*k6tv1.VirtualMachineInstance)
	if !ok {
		return
	}
	
	// check VMI status for Succeeded
	if vmi.Status.Phase != k6tv1.Succeeded {
		return
	}
	
	removeWarningLabel(vmi)	
}

func handleVmiDeletion(obj interface{}) {
	vmi, ok := obj.(*k6tv1.VirtualMachineInstance)
	if !ok {
		return
	}
	
	removeWarningLabel(vmi)
}

func removeWarningLabel(vmi *k6tv1.VirtualMachineInstance) {
	
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
