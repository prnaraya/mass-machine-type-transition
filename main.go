package mass_machine_type_transition

import (
	"fmt"
	"os"
)

func main() {
	virtCli, err := getVirtCli()
	if err != nil {
		os.Exit(1)
	}
	
	vmiInformer, err := getVmiInformer(virtCli)
	if err != nil {
		os.Exit(1)
	}
	
	exitJob = make(chan struct{})
	go vmiInformer.Run(exitJob)
	
	
	err = updateMachineTypes(virtCli)
	if err != nil {
		fmt.Println(err)
	}
	
	// wait for list of VMIs that need restart to be empty
	<-exitJob
	
	os.Exit(0)
}
