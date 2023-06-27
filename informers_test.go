package mass_machine_type_transition

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	k6tv1 "kubevirt.io/api/core/v1"
)

var _ = Describe("Informers", func() {
		
		Describe("handleDeleteVmi", func() {
			
			var	testVMI *k6tv1.VirtualMachineInstance
			
			BeforeEach(func() {
				ignoreKubeClientError = true
				testVMI = &k6tv1.VirtualMachineInstance {
					ObjectMeta: k8sv1.ObjectMeta{
						Name: "test-vm",
						Namespace: k8sv1.NamespaceDefault,
					},
				}
				
				vmKey, err := cache.MetaNamespaceKeyFunc(testVMI)
				Expect(err).ToNot(HaveOccurred())
				vmisPendingUpdate[vmKey] = struct{}{}
			})
				
			It("should remove VMI key from list of VMIs pending update", func() {
				testVMI2 := &k6tv1.VirtualMachineInstance {
					ObjectMeta: k8sv1.ObjectMeta{
						Name: "test-vm2",
						Namespace: k8sv1.NamespaceDefault,
					},
				}
				
				vmKey, err := cache.MetaNamespaceKeyFunc(testVMI2)
				Expect(err).ToNot(HaveOccurred())
				vmisPendingUpdate[vmKey] = struct{}{}
			
				handleDeletedVmi(testVMI2)
				Expect(vmisPendingUpdate).ToNot(HaveKey(vmKey))
				
				vmKey, err = cache.MetaNamespaceKeyFunc(testVMI)
				Expect(err).ToNot(HaveOccurred())
				delete(vmisPendingUpdate, vmKey)
			})
			
			When("VMI is final VMI in list of VMIs pending update", func() {
				It("should signal job to exit", func() {
					handleDeletedVmi(testVMI)
					Expect(vmisPendingUpdate).To(BeEmpty())
				})
			})
		})
})

func newVmWithLabel() *k6tv1.VirtualMachine {
	testVM := &k6tv1.VirtualMachine{
		ObjectMeta: k8sv1.ObjectMeta{
			Name: "test-vm",
			Namespace: k8sv1.NamespaceDefault,
			Labels: map[string]string{"restart-vm-required": "true"},
		},
	}
	return testVM
}
