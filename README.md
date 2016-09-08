# k8s-utils

Various one-off utilities useful to folks running kubernetes clusters.

## Developing

You'll need:

* a (recent) version of golang
* [gb](https://getgb.io/)

## Deploying

Kubernetes configurations to run the various k8s-utils can be found in
`config/kubernetes`. Simply `kube apply -f` the ones you want to use.

## Utils

### job-gc

Kubernetes Jobs are a great way to invoke batch jobs that are run-once in nature.

For Reasons, the pods they create are not cleaned up after exit.
Cleaning these up manually is somewhat annoying, so job-gc will poll every N
seconds for completed Jobs and delete their pods for you.

Right now this operates as a long-lived Deployment, but when recurring/scheduled
Job support is added in Kubernetes 1.4, we can use that instead.
