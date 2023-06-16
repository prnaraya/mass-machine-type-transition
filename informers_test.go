package mass_machine_type_transition

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	getVirtCli() test cases:
		- Successfully get VirtCli
		- Error getting clientConfig
		- Error getting VirtCli
	
	getVmiInformer() test case(s):
		- Successfully get VMI informer (one VMI)
		- VMI informer for each VMI (multiple VMIs)
		
	handleDeletedVMI() test cases:
		- vmiKey does not exist in list (nothing to remove)
		- vmiKey exists in list (remove from list)
		- vmiKey is last thing in list (list empty, exitJob channel should close)
*/
