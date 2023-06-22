package mass_machine_type_transition

import (
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k6tv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
//	"kubevirt.io/kubevirt/pkg/testutils"
)

/*
	updateMachineType() test cases:
		- VM machine type < rhel9.0.0 and not running
		- VM machine type < rhel9.0.0 and running, restartNow=false
		- VM machine type < rhel9.0.0 and running, restartNow=true
		- VM machine type >= rhel9.0.0
		- VM machine type format not pc-q35-rhelx.x.x
		- VM machine type format is default format (q35)
	
	addWarningLabel() test cases:
		- restartNow=false
		- restartNow=true (label will be removed basically as soon as it is applied since it is being restarted immediately)
*/

var _ = Describe("Update Machine Type", func() {
		var ctrl *gomock.Controller
		var virtClient *kubecli.MockKubevirtClient
		var vmiInterface *kubecli.MockVirtualMachineInstanceInterface
		var vmInterface *kubecli.MockVirtualMachineInterface
		
		BeforeEach(func() {
			//set up mock KubevirtClient and test VM
			ctrl = gomock.NewController(GinkgoT())
			virtClient = kubecli.NewMockKubevirtClient(ctrl)
			vmiInterface = kubecli.NewMockVirtualMachineInstanceInterface(ctrl)
			vmInterface = kubecli.NewMockVirtualMachineInterface(ctrl)
			virtClient.EXPECT().VirtualMachineInstance(k8sv1.NamespaceDefault).Return(vmiInterface).AnyTimes()
			virtClient.EXPECT().VirtualMachine(k8sv1.NamespaceDefault).Return(vmInterface).AnyTimes()
		})
	
		Describe("AddWarningLabel", func() {
			
/*			It("should apply warning label to VM", func() {
				
				
			})
*/				
			It("should add VM key to list of VMIs that must be restarted", func() {
				vm := NewVMWithMachineType("q35", false)
				addLabel := fmt.Sprint(`{"metadata": {"labels": {"restart-vm-required": "true"}}}}}`)
				vmInterface.EXPECT().Patch(vm.Name, types.StrategicMergePatchType, []byte(addLabel), &k8sv1.PatchOptions{}).Return(vm, nil)
				
				err := addWarningLabel(virtClient, vm)
				Expect(err).ToNot(HaveOccurred())
				
				vmKey := "default/test-vm-q35"
				_, inVmiList := vmisPendingUpdate[vmKey]
				Expect(inVmiList).To(BeTrue())
			})	
		})
		
		/*Describe("UpdateMachineType", func() {
			
			DescribeTable("when VM Machine Type is", func(machineType string) {
				//create mock VM with specified machine type
				//call 	
			},
				Entry("q35", "q35"),
				Entry("pc-q35-rhelx.x.x and less than latest machine type version", "pc-q35-rhel8.2.0"),
				Entry("pc-q35-rhelx.x.x and greater than or equal to latest machine type version", "pc-q35-rhel9.0.0"),
			)
		})*/
})

func NewVMWithMachineType(machineType string, running bool) *k6tv1.VirtualMachine {
	vmName := "test-vm-" + machineType
	testVM := &k6tv1.VirtualMachine{
		ObjectMeta: k8sv1.ObjectMeta{
			Name: vmName,
			Namespace: "default",
		},
		Spec: k6tv1.VirtualMachineSpec{
			Running: &running,
			Template: &k6tv1.VirtualMachineInstanceTemplateSpec{
				Spec: k6tv1.VirtualMachineInstanceSpec{
					Domain: k6tv1.DomainSpec{
						Machine: &k6tv1.Machine{
							Type: machineType,
						},
					},
				},
			},
		},
	}
	return testVM
}
