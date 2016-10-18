default:
	docker build -t k8s-utils-buildenv -f Dockerfile.buildenv .
	docker run --rm -v $(PWD):/app -w /app k8s-utils-buildenv:latest gb build

build-release: default
	docker build -t dubs/k8s-utils:$$(git rev-parse HEAD) -f Dockerfile.release .
	docker tag dubs/k8s-utils:$$(git rev-parse HEAD) dubs/k8s-utils:latest

release: build-release
	docker push dubs/k8s-utils:latest
	docker push dubs/k8s-utils:$$(git rev-parse HEAD)

kube:
	kubectl delete -f config/kubernetes || kubectl apply -f config/kubernetes

event-logger:
	kubectl logs "$$(kubectl get pods | grep event-logger | awk '{ print $$1; }')" --follow