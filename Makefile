default:
	docker build -t k8s-utils-buildenv -f Dockerfile.buildenv .
	docker run --rm -v $(PWD):/app -w /app k8s-utils-buildenv:latest gb build

build-release: default
	docker build -t dubs/k8s-utils -f Dockerfile.release .

release: build-release
	docker push dubs/k8s-utils

kube:
	kubectl delete -f config/kubernetes/job-gc.yml
	kubectl apply -f config/kubernetes/job-gc.yml
