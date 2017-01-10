
ifndef BUILDBASE
	export BUILDBASE=$(GOPATH)/src/github.com/crunchydata/crunchy-proxy
endif

default:
	make proxybin
gendeps:
	godep save \
	github.com/crunchydata/crunchy-proxy/proxy \
	github.com/crunchydata/crunchy-proxy/admin \
	github.com/crunchydata/crunchy-proxy/adapter \
	github.com/crunchydata/crunchy-proxy/config 

docsbuild:
	cd docs && ./build-docs.sh
clean:
	rm -rf $(GOPATH)/pkg/* $(GOPATH)/bin/*
	go get github.com/tools/godep
dockerimage:
	cp $(GOBIN)/crunchyproxy bin
	docker build -t crunchy-proxy -f Dockerfile.centos7 .
	docker tag crunchy-proxy crunchydata/crunchy-proxy:centos7-0.0.1
pushdockerimage:
	docker push crunchydata/crunchy-proxy:centos7-0.0.1

proxybin:
	godep go install crunchyproxy.go
all:
	make proxybin
push:
	./bin/push-to-dockerhub.sh

run:
	go run crunchyproxy.go -config=config.json
test:
	cd tests && go test; /usr/bin/test "$$?" -eq 0

