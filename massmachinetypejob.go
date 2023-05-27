package mass_machine_type_transition
/*
import (
	"context"
	//"strings"
	
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubevirt.io/client-go/kubecli"
)


// leftover from nested job approach
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
						Command: "",
					    },
					},
					RestartPolicy: v1.RestartPolicyOnFailure,
				},
			},
		},
	}
	
	_, err := jobs.Create(context.TODO(), jobSpec, k8sv1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}*/
