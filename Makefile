.DEFAULT_GOAL:=help
SHELL:=/bin/bash
OPERATOR_NAME=ghost-operator
KUBECONFIG?="$(HOME)/.kube/config"
NAMESPACE?=ghost
GO111MODULE=on
REGISTRY_NAME?=fossildev
IMAGE_NAME=$(OPERATOR_NAME)
IMAGE_VERSION?=0.0.1
IMAGE_TAG?=$(REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_VERSION)

##@ Installation

.PHONY: install
install: ## Install all resources (CRD's, RBAC and Operator)
	@echo ....... Creating namespace ....... 
	- kubectl create namespace ${NAMESPACE}
	@echo ....... Applying CRDs .......
	- kubectl apply -f deploy/crds/ghost.fossil.or.id_ghostapps_crd.yaml -n ${NAMESPACE}
	@echo ....... Applying Rules and Service Account .......
	- kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role_binding.yaml  -n ${NAMESPACE}
	- kubectl apply -f deploy/service_account.yaml  -n ${NAMESPACE}
	@echo ....... Applying Operator .......
	- kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}

.PHONY: uninstall
uninstall: ## Uninstall all that all performed in the $ make install.
	@echo ....... Uninstalling .......
	@echo ....... Deleting CRDs.......
	- kubectl delete -f deploy/crds/ghost.fossil.or.id_ghostapps_crd.yaml -n ${NAMESPACE}
	@echo ....... Deleting Rules and Service Account .......
	- kubectl delete -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}
	@echo ....... Deleting Operator .......
	- kubectl delete -f deploy/operator.yaml -n ${NAMESPACE}
	@echo ....... Deleting namespace ${NAMESPACE}.......
	- kubectl delete namespace ${NAMESPACE}

##@ Development

.PHONY: dep
dep: ## Update dependencies.
	@echo ....... Updating dependencies .......
	hack/update-dep.sh

.PHONY: code-gen
code-gen: ## Update generated code (k8s, openapi and crd).
	@echo ....... Updating generated code .......
	hack/update-codegen.sh

.PHONY: build
build: dep code-gen ## Build go binary and docker image.
	@echo ........ Building ghost operator with image tag ${IMAGE_TAG} ..........
	- operator-sdk build ${IMAGE_TAG}

.PHONY: publish
publish: ## Push docker image to container registry.
	@echo ........ Pushing ghost operator image ${IMAGE_TAG} ..........
	- docker push ${IMAGE_TAG}

##@ Tests

.PHONY: test
test: ## Run go unit tests.
	@echo ... Running unit test ...
	- go test -v ./pkg/...

.PHONY: test-e2e
test-e2e: ## Run integration e2e tests.
	@echo ... Running e2e test ...
	- operator-sdk test local ./test/e2e --namespace=${NAMESPACE} --image=${IMAGE_TAG} --kubeconfig=${KUBECONFIG} --verbose

.PHONY: test-e2e-local
test-e2e-local: ## Run integration e2e tests on local machine.
	@echo ... Running e2e test ...
	- operator-sdk test local ./test/e2e --up-local --namespace=${NAMESPACE} --image=${IMAGE_TAG} --kubeconfig=${KUBECONFIG} --verbose

.PHONY: help
help: ## Display this help.
	@echo -e "Usage:\n  make \033[36m<target>\033[0m"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)