package mass_machine_type_transition

import "os"

func main() {
	virtCli, err := getVirtCli()
	if err != nil {
		os.Exit(1)
	}
	
	vmiInformer, err := getVmiInformer(virtCli)
	if err != nil {
		os.Exit(1)
	}
	
	stopCh := make(chan struct{})
	go vmiInformer.Run(stopCh)
	
	updateMachineTypes(virtCli)
	
	// for now this uses the number of vms to be restarted to determine
	// when to exit the job rather than seeing if the actual label exists
	// on any VMs before exiting the job
	
	for (needsRestart > 0) {
	}
	
	os.Exit(0)
}
