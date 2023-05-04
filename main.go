package mass_machine_type_transition

import "os"

func main() {
	virtCli, err := getVirtCli()
	if err != nil {
		os.Exit(1)
	}
	
	/*vmiInformer, err := getVmiInformer(virtCli)
	if err != nil {
		os.Exit(1)
	}

	if vmiInformer != nil {
		os.Exit(1)
	}*/

	updateMachineTypes(virtCli)
	os.Exit(0)
}
