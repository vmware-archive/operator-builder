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

define create_path
	if [ ! -d $$1 ]; then\
        mkdir -p $$1;\
    fi
endef

export BASE_DIR := $(shell pwd)
export PATH := $(PATH):$(BASE_DIR):$(BASE_DIR)/bin:/usr/local/bin

build:
	go build -o bin/operator-builder cmd/operator-builder/main.go

install:
	sudo cp bin/operator-builder /usr/local/bin/operator-builder

#
# traditional testing
#
test:
	go test -cover -coverprofile=./bin/coverage.out ./...

test-coverage-view: test-install
	go tool cover -html=./bin/coverage.out	

#
# debug testing with delve
#
TEST_WORKLOAD_PATH ?= test/application

debug-clean:
	rm -rf $(TEST_WORKLOAD_PATH)/*

debug-init: debug-clean
	dlv debug ./cmd/operator-builder --wd $(TEST_WORKLOAD_PATH) -- $(INIT_OPTS)

debug-create:
	dlv debug ./cmd/operator-builder --wd $(TEST_WORKLOAD_PATH) -- $(CREATE_OPTS)

debug: debug-init debug-create

#
# simple generation testing outside of codebase itself
#
TEST_PATH ?= /tmp/test

generate-clean:
	if [ -d $(TEST_PATH) ]; then rm -rf $(TEST_PATH)/*; fi

generate-init: build generate-clean
	$(call create_path $(TEST_PATH))
	cp -r $(BASE_DIR)/$(TEST_WORKLOAD_PATH)/.workloadConfig $(TEST_PATH) ;
	ls -altr $(TEST_PATH) ;
	ls -altr $(TEST_PATH)/.workloadConfig ;
	cd $(TEST_PATH) && operator-builder $(INIT_OPTS)

generate-create:
	cd $(TEST_PATH) && operator-builder $(CREATE_OPTS)

generate: generate-init generate-create
