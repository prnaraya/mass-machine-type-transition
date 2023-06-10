package mass_machine_type_transition

import "os"

func main() {
	addFlags()
	flag.Parse()
	
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
	
	// wait for list of VMIs that need restart to be empty
	<-exitJob
	
	os.Exit(0)
}
