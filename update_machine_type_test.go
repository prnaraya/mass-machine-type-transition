package mass_machine_type_transition

import (
	"fmt"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	k6tv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

var _ = Describe("Update Machine Type", func() {
		var ctrl *gomock.Controller
		var virtClient *kubecli.MockKubevirtClient
		var vmInterface *kubecli.MockVirtualMachineInterface
		
		BeforeEach(func() {
			ignoreKubeClientError = false
			ctrl = gomock.NewController(GinkgoT())
			virtClient = kubecli.NewMockKubevirtClient(ctrl)
			vmInterface = kubecli.NewMockVirtualMachineInterface(ctrl)
			virtClient.EXPECT().VirtualMachine(k8sv1.NamespaceDefault).Return(vmInterface).AnyTimes()
			vmInterface.EXPECT().List(gomock.Any()).AnyTimes()
		})
		
		Describe("addWarningLabel", func() {

			It("should add VM Key to list of VMIs that need to be restarted", func() {
				vm := newVMWithMachineType("q35", true)
				vm.Labels = map[string]string{}
				
				vmInterface.EXPECT().Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(vmName string, pt types.PatchType, data []byte, patchOpts *k8sv1.PatchOptions, subresources ...string) {
					vm.Labels["restart-vm-required"] = "true"
				}).AnyTimes()
				
				err := addWarningLabel(virtClient, vm)
				Expect(err).ToNot(HaveOccurred())
				vmKey, err := cache.MetaNamespaceKeyFunc(vm)
				Expect(err).ToNot(HaveOccurred())
				Expect(vmisPendingUpdate).To(HaveKey(vmKey))

				Expect(vm.Labels).To(HaveKeyWithValue("restart-vm-required", "true"), "VM should have 'restart-vm-required' label")
				
				delete(vmisPendingUpdate, vmKey)
			})
		})
		
		Describe("patchVmMachineType", func() {
		
			DescribeTable("when machine type is", func(machineType string) {
				vm := newVMWithMachineType(machineType, false)
				updateMachineTypeVersion, parsedMachineType := getMachineTypeVersion(machineType)
				updateMachineType := fmt.Sprintf(`{"spec": {"template": {"spec": {"domain": {"machine": {"type": "%s"}}}}}}`, updateMachineTypeVersion)

				vmInterface.EXPECT().Patch(vm.Name, types.StrategicMergePatchType, []byte(updateMachineType), &k8sv1.PatchOptions{}).Do(func(vmName string, pt types.PatchType, data []byte, patchOpts *k8sv1.PatchOptions, subresources ...string) {
					if parsedMachineType == "q35" || parsedMachineType < latestMachineTypeVersion {
						vm.Spec.Template.Spec.Domain.Machine.Type = updateMachineTypeVersion
					}
				}).AnyTimes()
				
				err := patchVmMachineType(virtClient, vm, machineType)
				Expect(err).ToNot(HaveOccurred())
				
				if parsedMachineType >= minimumSupportedMachineTypeVersion {
					Expect(vm.Spec.Template.Spec.Domain.Machine.Type).To(Equal(machineType))
				} else {
					Expect(vm.Spec.Template.Spec.Domain.Machine.Type).To(Equal(updateMachineTypeVersion))
				}
				
				vmKey, err := cache.MetaNamespaceKeyFunc(vm)
				delete(vmisPendingUpdate, vmKey)
			},
				Entry("'q35' should update machine type to latest version", "q35"),
				Entry("'pc-q35-rhelx.x.x' and less than minimum supported machine type version should update machine type to latest version", "pc-q35-rhel8.2.0"),
				Entry("'pc-q35-rhelx.x.x' and greater than or equal to latest machine type version should not affect machine type", "pc-q35-rhel9.0.0"),
			)
		})
})

func newVMWithMachineType(machineType string, running bool) *k6tv1.VirtualMachine {
	vmName := "test-vm-" + machineType
	testVM := &k6tv1.VirtualMachine{
		ObjectMeta: k8sv1.ObjectMeta{
			Name: vmName,
			Namespace: k8sv1.NamespaceDefault,
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
	testVM.Labels = map[string]string{}
	return testVM
}

func getMachineTypeVersion(machineType string) (string, string) {
	parsedMachineType := "q35"
	if machineType != "q35" {
		splitMachineType := strings.Split(machineType, "-")
		parsedMachineType = splitMachineType[2]
	}
	updateMachineTypeVersion := "pc-q35-" + latestMachineTypeVersion
	return updateMachineTypeVersion, parsedMachineType
}
