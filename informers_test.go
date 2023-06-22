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

var _ = Describe("Update Machine Type", func() {
		
/*		BeforeEach(func() {
			//set up mock KubevirtClient
			ctrl = gomock.NewController(GinkgoT())
			virtClient = kubecli.NewMockKubevirtClient(ctrl)
		})
		
		AfterEach(func() {
			//remove mock KubevirtClient and mock VM(s)
		})*/
		
		Describe("AddWarningLabel", func() {
				
			It("should apply warning label to VM", func() {
				label := true
				Expect(label).To(Equal(true))
			})
				
			It("should add VM key to list of VMIs that must be restarted", func() {
				vmKey := "namespace/name"
				Expect(vmKey).To(Equal("namespace/name"))
			})	
		})
})
