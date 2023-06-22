# Mass machine type transition
This project contains the k8s job that will perform a mass machine type transition on an OpenShift cluster.

## Notes
This is a more cohesive version of the mass machine type transition. This implementation uses Docker to create an image from the golang script that performs the machine type transition, and hangs until all warnings are cleared. The .yaml file is a job that uses this generated image, and will only be completed once this script exits, i.e. there are no warning labels left on any VMs. To run this locally, you will need to build a Docker image with the tag "machine-type-transition", then start the job with kubectl apply. This is not comprehensively tested yet; in the future I will be automating the process of running this job, as well as adding functionality for the user to specify flags to update machine types only on specific namespaces, and to force a restart/shut down of all running VMs when updating their machine types.
