# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
INIT_OPTS=init \
		--workload-config .workloadConfig/workload.yaml \
   		--repo github.com/acme/acme-cnp-mgr \
        --skip-go-version-check
CREATE_OPTS=create api \
	--workload-config .workloadConfig/workload.yaml \
	--controller \
	--resource

set-path:
	export PATH=$$PATH:$$PWD:$$PWD/bin

build:
	go build -o bin/operator-builder cmd/operator-builder/main.go

#
# traditional testing
# TODO: come to consensus on what this looks like versus debug/generate tests
#
TEST_PATH ?= /tmp
TEST_SCRIPT ?= default.sh

test-install: build
	go test -cover -coverprofile=./bin/coverage.out ./...

test-coverage-view: test-install
	go tool cover -html=./bin/coverage.out	

test: test-install set-path
	find . -name ${TEST_SCRIPT} | xargs dirname | xargs -I {} cp -r {} $(TEST_PATH)/.workloadConfig
	cd $(TEST_PATH); basename ${TEST_SCRIPT} | xargs find ${TEST_PATH} -name | xargs sh

#
# debug testing with delve
#
DEBUG_PATH ?= test/application

debug-clean:
	rm -rf $(DEBUG_PATH)/*

debug-init: debug-clean
	dlv debug ./cmd/operator-builder --wd $(DEBUG_PATH) -- $(INIT_OPTS)

debug-create:
	dlv debug ./cmd/operator-builder --wd $(DEBUG_PATH) -- $(CREATE_OPTS)

debug: debug-init debug-create

#
# simple generation testing
#
GENERATE_PATH ?= $(DEBUG_PATH)

generate-clean:
	rm -rf $(GENERATE_PATH)/*

generate-init: test-clean set-path
	cd $(GENERATE_PATH) && operator-builder $(INIT_OPTS)

generate-create: set-path
	cd $(GENERATE_PATH) && operator-builder $(CREATE_OPTS)

generate: generate-init generate-create
