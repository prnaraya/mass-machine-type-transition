package mass_machine_type_transition

import (
	"context"
	//"strings"
	
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/client-go/kubecli"
)

func calculateCompletionCount(virtCli kubecli.KubevirtClient, namespace string) (int, error) {
	completionCount := 0
	vmList, err := virtCli.VirtualMachine(namespace).List(&k8sv1.ListOptions{})
	if err != nil {
		return completionCount, err
	}
	for _, _ = range vmList.Items {
		// we really only want to update machine types that are before rhel9, so we need a way to filter those out both when getting
		// the number of completions, and when running the job on nodes, presumably by adding some kind of label to mark the nodes as out of date,
		// then using the node selector in the job spec to only run on the specified nodes
		
		/*machineType := vm.Spec.Template.Domain.Machine.Type
		machineTypeSubstrings := strings.Split(machineType, "-")
		
		if strings.Contains(machineType, "rhel") {
			
		}*/
		completionCount++
	}
	return completionCount, nil
}

func generateMachineTypeTransitionJob(virtCli kubecli.KubevirtClient, namespace *string, completionCount *int32) error {
	jobs := virtCli.BatchV1().Jobs("default")
	jobSpec := &batchv1.Job{
		ObjectMeta: k8sv1.ObjectMeta{
			Name:      "mass-machine-type-transition",
			Namespace: *namespace,
		},
		Spec: batchv1.JobSpec{
			Completions: completionCount,
			Template:    v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
					    {
						Name:    "mass-machine-type-transition",
						// not 100% sure what to use for image here
						Image:   "",
						Command: []string{"kubectl"},
						// currently this will modify ALL vm's to have this specific machine type, however this will be expanded to accommodate all
						// supported RHEL machine types as listed by qemu-kvm
						Args:    []string{"patch", "virtualmachine", "testvm", "--type", "merge", "--patch", `{"spec":{"template":{"spec":{"domain":{"machine":{"type": "pc-q35-rhel9.0.0"}}}}}}`},
					    },
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
	
	_, err := jobs.Create(context.TODO(), jobSpec, k8sv1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
