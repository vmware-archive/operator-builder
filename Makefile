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

build:
	go build -o bin/operator-builder cmd/operator-builder/main.go

test-install: build
	go test -cover -coverprofile=./bin/coverage.out ./...
	sudo cp bin/operator-builder /usr/local/bin/operator-builder

test-coverage-view: test-install
	go tool cover -html=./bin/coverage.out	

DEBUG_PATH ?= test/application
TEST_PATH ?= $(DEBUG_PATH)

debug-clean:
	rm -rf $(DEBUG_PATH)/*

debug-init: debug-clean
	dlv debug ./cmd/operator-builder --wd $(DEBUG_PATH) -- $(INIT_OPTS)

debug-create:
	dlv debug ./cmd/operator-builder --wd $(DEBUG_PATH) -- $(CREATE_OPTS)

debug: debug-init debug-create

test-clean:
	rm -rf $(TEST_PATH)/*

test-init: test-clean
	cd $(TEST_PATH) && $$OLDPWD/bin/operator-builder $(INIT_OPTS)

test-create:
	cd $(TEST_PATH) && $$OLDPWD/bin/operator-builder $(CREATE_OPTS)
