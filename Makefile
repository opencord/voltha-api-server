#
# Copyright 2016 the original author or authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# set default shell
SHELL = bash -e -o pipefail

# Variables
VERSION                    ?= $(shell cat ./VERSION)

DOCKER_LABEL_VCS_DIRTY     = false
ifneq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
    DOCKER_LABEL_VCS_DIRTY = true
endif
## Docker related
DOCKER_EXTRA_ARGS          ?=
DOCKER_REGISTRY            ?=
DOCKER_REPOSITORY          ?=
DOCKER_TAG                 ?= ${VERSION}$(shell [[ ${DOCKER_LABEL_VCS_DIRTY} == "true" ]] && echo "-dirty" || true)
AFROUTER_IMAGENAME         := ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}voltha-afrouter
AFROUTERD_IMAGENAME        := ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}voltha-afrouterd

## Docker labels. Only set ref and commit date if committed
DOCKER_LABEL_VCS_URL       ?= $(shell git remote get-url $(shell git remote))
DOCKER_LABEL_VCS_REF       = $(shell git rev-parse HEAD)
DOCKER_LABEL_BUILD_DATE    ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
DOCKER_LABEL_COMMIT_DATE   = $(shell git show -s --format=%cd --date=iso-strict HEAD)

DOCKER_BUILD_ARGS ?= \
	${DOCKER_EXTRA_ARGS} \
	--build-arg org_label_schema_version="${VERSION}" \
	--build-arg org_label_schema_vcs_url="${DOCKER_LABEL_VCS_URL}" \
	--build-arg org_label_schema_vcs_ref="${DOCKER_LABEL_VCS_REF}" \
	--build-arg org_label_schema_build_date="${DOCKER_LABEL_BUILD_DATE}" \
	--build-arg org_opencord_vcs_commit_date="${DOCKER_LABEL_COMMIT_DATE}" \
	--build-arg org_opencord_vcs_dirty="${DOCKER_LABEL_VCS_DIRTY}"

DOCKER_BUILD_ARGS_LOCAL ?= ${DOCKER_BUILD_ARGS} \
	--build-arg LOCAL_PROTOS=${LOCAL_PROTOS}

# Default is GO111MODULE=auto, which will refuse to use go mod if running
# go less than 1.13.0 and this repository is checked out in GOPATH. For now,
# force module usage. This affects commands executed from this Makefile, but
# not the environment inside the Docker build (which does not build from
# inside a GOPATH).
export GO111MODULE=on

.PHONY: rw_core ro_core afrouter afrouterd local-protos local-lib-go

# This should to be the first and default target in this Makefile
help:
	@echo "Usage: make [<target>]"
	@echo "where available targets are:"
	@echo
	@echo "build                : Build the docker images."
	@echo "                         - If this is the first time you are building, choose 'make build' option."
	@echo "afrouter             : Build the afrouter docker image"
	@echo "afrouterd            : Build the afrouterd docker image"
	@echo "clean                : Remove files created by the build and tests"
	@echo "distclean            : Remove venv directory"
	@echo "docker-push          : Push the docker images to an external repository"
	@echo "lint-dockerfile      : Perform static analysis on Dockerfiles"
	@echo "lint-style           : Verify code is properly gofmt-ed"
	@echo "lint-sanity          : Verify that 'go vet' doesn't report any issues"
	@echo "lint-mod             : Verify the integrity of the 'mod' files"
	@echo "lint                 : Shorthand for lint-style & lint-sanity"
	@echo "local-lib-go         : Copies a local version of the VOTLHA dependencies into the vendor directory"
	@echo "local-protos         : Copies a local verison of the VOLTHA protos into the vendor directory"
  @echo "sca                  : Runs various SCA through golangci-lint tool"
  @echo "test                 : Generate reports for all go tests"
	@echo

## Local Development Helpers
local-protos:
ifdef LOCAL_PROTOS
	rm -rf vendor/github.com/opencord/voltha-protos/go
	mkdir -p vendor/github.com/opencord/voltha-protos/go
	cp -r ${LOCAL_PROTOS}/go/* vendor/github.com/opencord/voltha-protos/go
	rm -rf vendor/github.com/opencord/voltha-protos/go/vendor
endif

## Local Development Helpers
local-lib-go:
ifdef LOCAL_LIB_GO
	rm -rf vendor/github.com/opencord/voltha-lib-go
	mkdir -p vendor/github.com/opencord/voltha-lib-go/v2/pkg
	cp -r ${LOCAL_LIB_GO}/pkg/* vendor/github.com/opencord/voltha-lib-go/v2/pkg/
endif

## Docker targets

build: docker-build

docker-build: afrouter afrouterd

afrouter: local-protos local-lib-go
	docker build $(DOCKER_BUILD_ARGS) -t ${AFROUTER_IMAGENAME}:${DOCKER_TAG} -t ${AFROUTER_IMAGENAME}:latest -f docker/Dockerfile.afrouter .

afrouterd: local-protos local-lib-go
	docker build $(DOCKER_BUILD_ARGS) -t ${AFROUTERD_IMAGENAME}:${DOCKER_TAG} -t ${AFROUTERD_IMAGENAME}:latest -f docker/Dockerfile.afrouterd .

docker-push:
	docker push ${AFROUTER_IMAGENAME}:${DOCKER_TAG}
	docker push ${AFROUTERD_IMAGENAME}:${DOCKER_TAG}

## lint and unit tests

PATH:=$(GOPATH)/bin:$(PATH)
HADOLINT=$(shell PATH=$(GOPATH):$(PATH) which hadolint)
lint-dockerfile:
ifeq (,$(shell PATH=$(GOPATH):$(PATH) which hadolint))
	mkdir -p $(GOPATH)/bin
	curl -o $(GOPATH)/bin/hadolint -sNSL https://github.com/hadolint/hadolint/releases/download/v1.17.1/hadolint-$(shell uname -s)-$(shell uname -m)
	chmod 755 $(GOPATH)/bin/hadolint
endif
	@echo "Running Dockerfile lint check ..."
	@hadolint $$(find . -name "Dockerfile.*")
	@echo "Dockerfile lint check OK"

lint-style:
ifeq (,$(shell which gofmt))
	go get -u github.com/golang/go/src/cmd/gofmt
endif
	@echo "Running style check..."
	@gofmt_out="$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" ;\
	if [ ! -z "$$gofmt_out" ]; then \
	  echo "$$gofmt_out" ;\
	  echo "Style check failed on one or more files ^, run 'go fmt' to fix." ;\
	  exit 1 ;\
	fi
	@echo "Style check OK"

lint-sanity:
	@echo "Running sanity check..."
	@go vet -mod=vendor ./...
	@echo "Sanity check OK"

lint-mod:
	@echo "Running dependency check..."
	@go mod verify
	@echo "Dependency check OK"

lint: lint-style lint-sanity lint-mod lint-dockerfile

# Rules to automatically install golangci-lint
GOLANGCI_LINT_TOOL?=$(shell which golangci-lint)
ifeq (,$(GOLANGCI_LINT_TOOL))
GOLANGCI_LINT_TOOL=$(GOPATH)/bin/golangci-lint
golangci_lint_tool_install:
	# Same version as installed by Jenkins ci-management
	# Note that install using `go get` is not recommended as per https://github.com/golangci/golangci-lint
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOPATH)/bin v1.17.0
else
golangci_lint_tool_install:
endif

# Rules to automatically install go-junit-report
GO_JUNIT_REPORT?=$(shell which go-junit-report)
ifeq (,$(GO_JUNIT_REPORT))
GO_JUNIT_REPORT=$(GOPATH)/bin/go-junit-report
go_junit_install:
	go get -u github.com/jstemmer/go-junit-report
else
go_junit_install:
endif

# Rules to automatically install gocover-covertura
GOCOVER_COBERTURA?=$(shell which gocover-cobertura)
ifeq (,$(GOCOVER_COBERTURA))
	@GOCOVER_COBERTURA=$(GOPATH)/bin/gocover-cobertura
gocover_cobertura_install:
	go get -u github.com/t-yuki/gocover-cobertura
else
gocover_cobertura_install:
endif

sca: golangci_lint_tool_install
	rm -rf ./sca-report
	@mkdir -p ./sca-report
	$(GOLANGCI_LINT_TOOL) run --deadline 10m --out-format junit-xml ./... 2>&1 | tee ./sca-report/sca-report.xml

test: go_junit_install gocover_cobertura_install
	@mkdir -p ./tests/results
	@go test -mod=vendor -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	$(GO_JUNIT_REPORT) < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	$(GOCOVER_COBERTURA) < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

clean:

distclean: clean
	rm -rf ./sca_report

# end file
