all: build
PACKAGE=github.com/bloomberg/k8eraid

ARCH?=amd64
GOLANG_VERSION?=1.12.0
CONTAINER_BUILD_IMAGE?=golang:$(GOLANG_VERSION)
REPO_DIR:=$(shell pwd)
GOPATH?=$(shell go env GOPATH)
DOCKER_RUN=docker run --rm -i $(TTY) -v $(TEMP_DIR):/build -v $(REPO_DIR):/go/src/$(PACKAGE):z -w /go/src/$(PACKAGE) -e GOARCH=$(ARCH)
DOCKER_IMAGE?=bloomberg/k8eraid
APPVERSION?=v0.8.1
SCRATCH_IMAGE?=scratch
SCRATCH_TAG?=""
DEP=$(GOPATH)/bin/dep

ifneq ("$(http_proxy)", "")
PROXY_VARS=http_proxy=$(http_proxy) https_proxy=$(http_proxy)
DOCKER_RUN += -e http_proxy=$(http_proxy) $(CONTAINER_BUILD_IMAGE)
gitconfig:
	git config http.proxy $(http_proxy)
	git config https.proxy $(http_proxy)
	git config url.https://github.com/.insteadof git://github.com/
else
DOCKER_RUN += $(CONTAINER_BUILD_IMAGE)
gitconfig:
	@echo no gitconfig
endif

ifndef TEMP_DIR
TEMP_DIR:=$(shell mktemp -d /tmp/k8eraid.XXXXXX)
endif

TTY=
ifeq ($(shell [ -t 0 ] && echo 1 || echo 0), 1)
	TTY=-t
endif

$(DEP):
	$(PROXY_VARS) go get github.com/golang/dep/cmd/dep

vendor: $(DEP)
	$(PROXY_VARS) $(DEP) ensure

build/k8eraid: clean vendor
	GOARCH=$(ARCH) go build -o build/k8eraid $(PACKAGE)/cmd/k8eraid

container: 
	# Run the build in a container in order to have reproducible builds
	$(DOCKER_RUN) make build/k8eraid
	docker build . --pull -t $(DOCKER_IMAGE):$(APPVERSION) -t $(DOCKER_IMAGE):latest --build-arg IMAGE=$(SCRATCH_IMAGE) --build-arg TAG=$(SCRATCH_TAG)
	docker build . -f Dockerfile.vendor -t $(DOCKER_IMAGE):$(APPVERSION)-vendor -t $(DOCKER_IMAGE):latest-vendor --build-arg IMAGE=$(SCRATCH_IMAGE) --build-arg TAG=$(SCRATCH_TAG)

pushcontainer:
	docker push $(DOCKER_IMAGE):$(APPVERSION)
	docker push $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(APPVERSION)-vendor
	docker push $(DOCKER_IMAGE):latest-vendor
	docker rmi $(DOCKER_IMAGE):$(APPVERSION)
	docker rmi $(DOCKER_IMAGE):latest
	docker rmi $(DOCKER_IMAGE):$(APPVERSION)-vendor
	docker rmi $(DOCKER_IMAGE):latest-vendor

test: clean vendor
	CGO_ENABLED=1 go test -race -v --cover ./...

clean:
	rm -rf build

gofmt:
	@hack/gofmt.sh

lint: clean
	CGO_ENABLED=0 golint -set_exit_status $(shell go list ./...)

.PHONY: all build gofmt lint lintcontainer pushcontainer testcontainer container clean test
