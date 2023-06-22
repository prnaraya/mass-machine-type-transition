package mass_machine_type_transition

import (
	"fmt"
	"strings"
	
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	
	// by default, update machine type across all namespaces
	namespace = k8sv1.NamespaceAll
	//labels []string
	
	// by default, should require manual restarting of VMIs
	restartNow = false
)

func patchVmMachineType(virtCli kubecli.KubevirtClient, vm *k6tv1.VirtualMachine, machineType string) error {
	updateMachineType := fmt.Sprintf(`{"spec": {"template": {"spec": {"domain": {"machine": {"type": "%s"}}}}}}`, machineType)
		
	_, err := virtCli.VirtualMachine(vm.Namespace).Patch(vm.Name, types.StrategicMergePatchType, []byte(updateMachineType), &k8sv1.PatchOptions{})
	if err != nil {
		return err
	}
	
	// add label to running VMs that a restart is required for change to take place
	if vm.Status.Created {
		// adding the warning label to the VMs regardless if we restart them now or if the user does it manually
		// shouldn't matter, since the deletion of the VMI will remove the label and remove the vmi list anyway
		err = addWarningLabel(virtCli, vm)
		if err != nil {
			return err
		}
		
		if restartNow {
			err = virtCli.VirtualMachine(vm.Namespace).Restart(vm.Name, &k6tv1.RestartOptions{})
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

func updateMachineTypes(virtCli kubecli.KubevirtClient) error {
	vmList, err := virtCli.VirtualMachine(namespace).List(&k8sv1.ListOptions{})
	if err != nil {
		return err
	}
	for _, vm := range vmList.Items {
		machineType := vm.Spec.Template.Spec.Domain.Machine.Type
		machineTypeSubstrings := strings.Split(machineType, "-")
		
		// in the case where q35 is the machine type, the VMI status will contain the
		// full machine type string. Since q35 is an alias for the most recent machine type,
		// if a VM hasn't been restarted and thus still has an outdated machine type, it should
		// be updated to the most recent machine type. I'm not able to access the machine type from
		// VMI status using the kv library, so instead I am opting to update the machine types of all
		// VMs with q35 instead. It might only be necessary to update the ones that are running as well,
		// though I am not sure.
		
		if len(machineTypeSubstrings) == 1 {
			if machineTypeSubstrings[0] == "q35" {
				machineType = "pc-q35-" + latestMachineTypeVersion
				return patchVmMachineType(virtCli, &vm, machineType)
			}
		}
		
		if len(machineTypeSubstrings) == 3 {
			version := machineTypeSubstrings[2]
			if strings.Contains(version, "rhel") && version < "rhel9.0.0" {
				machineTypeSubstrings[2] = latestMachineTypeVersion
				machineType = strings.Join(machineTypeSubstrings, "-")
				return patchVmMachineType(virtCli, &vm, machineType)
			}
		}
	}
	return nil
}

func addWarningLabel (virtCli kubecli.KubevirtClient, vm *k6tv1.VirtualMachine) error {
	addLabel := fmt.Sprint(`{"metadata": {"labels": {"restart-vm-required": "true"}}}}}`)
	_, err := virtCli.VirtualMachine(vm.Namespace).Patch(vm.Name, types.StrategicMergePatchType, []byte(addLabel), &k8sv1.PatchOptions{})
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
