
ifndef BUILDBASE
	export BUILDBASE=$(GOPATH)/src/github.com/crunchydata/crunchy-proxy
endif

gendeps:
	godep save \
	github.com/crunchydata/crunchy-proxy/proxy \
	github.com/crunchydata/crunchy-proxy/admin \
	github.com/crunchydata/crunchy-proxy/adapter \
	github.com/crunchydata/crunchy-proxy/config 

docs:
	cd docs && ./build-docs.sh
clean:
	rm -rf $(GOPATH)/pkg/* $(GOPATH)/bin/*
	go get github.com/tools/godep
image:
	docker build -t crunchy-proxy -f centos7/Dockerfile .
	docker tag crunchy-proxy crunchydata/crunchy-proxy:centos7-0.0.1

proxybin:
	godep go install crunchyproxy.go
all:
	make proxybin
push:
	./bin/push-to-dockerhub.sh
default:
	make proxybin
test:
	cd tests && go test; /usr/bin/test "$$?" -eq 0

