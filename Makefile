# Copyright 2017 Crunchy Data Solutions, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
#
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
# License for the specific language governing permissions and limitations under
# the License.

.PHONY: all build clean clean-docs docs resolve install release run default

all: clean resolve build

RELEASE_VERSION := v1.0.0beta
PROJECT_DIR := $(shell pwd)
BUILD_DIR := $(PROJECT_DIR)/build
DIST_DIR := $(PROJECT_DIR)/dist
VENDOR_DIR := $(PROJECT_DIR)/vendor
DOCS_DIR := $(PROJECT_DIR)/docs

BUILD_TARGET := $(PROJECT_DIR)/main.go

CRUNCHY_PROXY := crunchy-proxy
RELEASE_ARCHIVE := $(DIST_DIR)/crunchyproxy-$(RELEASE_VERSION).tar.gz


clean:
	@echo "Cleaning project..."
	@rm -rf $(DIST_DIR)
	@rm -rf $(VENDOR_DIR)
	@go clean -i

resolve:
	@echo "Resolving depenencies..."
	@glide up

build:
	@echo "Building project..."
	@go build -i -o $(BUILD_DIR)/$(CRUNCHY_PROXY)

install:
	@go install

clean-docs:
	@rm -rf $(DOCS_DIR)/pdf

docs:
	@mkdir -p $(DOCS_DIR)/pdf
	@cd docs && ./build-docs.sh

release: clean resolve build
	@echo "Creating $(RELEASE_VERSION) release archive..."
	@mkdir -p $(DIST_DIR)
	@tar czf $(RELEASE_ARCHIVE) -C $(BUILD_DIR) $(CRUNCHY_PROXY)
	@echo "Created: $(RELEASE_ARCHIVE)"

run:
	@go run main.go --config=./examples/config.yaml

#dockerimage:
#	cp $(GOBIN)/crunchyproxy bin
#	docker build -t crunchy-proxy -f Dockerfile.centos7
#	docker tag crunchy-proxy crunchydata/crunchy-proxy:centos7-$(PROXY_RELEASE)

#pushdockerimage:
#	docker push crunchydata/crunchy-proxy:centos7-$(PROXY_RELEASE)

#push:
#	./bin/push-to-dockerhub.sh

#test:
#	cd tests && go test; /usr/bin/test "$$?" -eq 0

