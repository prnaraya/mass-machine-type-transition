package mass_machine_type_transition

import "os"

func main() {
	vmiInformer, err := getVmiInformer()
	if err != nil {
		os.Exit(1)
	}

	if vmiInformer != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
