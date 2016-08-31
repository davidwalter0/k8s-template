.PHONY: deps
export GOPATH=/go
SHELL=/bin/bash
MAKEFILE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CURRENT_DIR := $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))
DIR=$(MAKEFILE_DIR)
# export HOSTNAME=$(shell hostname)

targets=bin/k8s-template

build: $(targets)

all: 
	@printf "\n--------------------------------\n"
	@printf "Running in abs directory\n    $(MAKEFILE_DIR)\n"
	@printf "The subdirectory is $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))\n"
	@printf "\n--------------------------------\n"
	@printf "make targets init to initialize godeps, get, save, test and build\n"

.dep: $(targets)
	touch .dep

%: bin/%

bin/%: %.go 
	@echo "Building via % rule for $@ from $<"

	@if go version|grep -q 1.4 ; then											\
	    args="-s -w -X main.Build $$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit $$(git log --format=%hash-%aI -n1)";	\
	fi;															\
	if go version|grep -qE "1\.[5-9]|2.*]" ; then											\
	    args="-s -w -X main.Build=$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit=$$(git log --format=%hash-%aI -n1)";	\
	fi;															\
	CGO_ENABLED=0 go get --tags netgo -ldflags "$${args}" ;									\
	CGO_ENABLED=0 go build --tags netgo -ldflags "$${args}" -o $@ $^ ;
	cp $@ /go/bin/

init: get save

get: 
	@GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH} $(GOPATH)/bin/godep get $(libdep)
save:
	@GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH} $(GOPATH)/bin/godep save
clean:
	@echo cleaning up temporary files
	@echo rm -f $(targets)
	@rm -f $(targets) .dep $(package_dir)/.dep

export SIMPLE_FILE_SERVER_PORT=65531
export SIMPLE_FILE_SERVER_PATH=$(MAKEFILE_DIR)
export SIMPLE_FILE_SERVER_HOST=localhost
export SIMPLE_FILE_SERVER=$${GOPATH}/bin/k8s-simple-file-server
export HOSTNAME=$(shell hostname)

pidfile=tests/pid
.PHONY: ${pidfile} pre-test run-test post-test 

pre-test: tests/pid
	set | grep -i hostname
	@echo ">SIMPLE_FILE_SERVER=$(SIMPLE_FILE_SERVER)<"
	@echo ">SIMPLE_FILE_SERVER_HOST=$(SIMPLE_FILE_SERVER_HOST)<"
	@echo ">SIMPLE_FILE_SERVER_PATH=$(SIMPLE_FILE_SERVER_PATH)<"
	@echo ">SIMPLE_FILE_SERVER_PORT=$(SIMPLE_FILE_SERVER_PORT)<"
	${SIMPLE_FILE_SERVER} &	\
	echo $$! > $<

post-test:
	ps -ef|grep ${SIMPLE_FILE_SERVER}
	pid=$$(cat ${pidfile}); ps -ef | grep ^$${pid}; kill -9 $${pid}; rm -f ${pidfile}

test: pre-test run-test post-test

run-test:
	@echo preprocess: defers file replacement in mappings
	K8SNameSpace=smoke bin/k8s-template --preprocess < tests/mappings.yaml > pre.yaml
	@echo replacement: replace text and file, uri mappings
	K8SNameSpace=smoke bin/k8s-template --mappings=tests/mappings.yaml --template=tests/unmap.txt
	K8SNameSpace=smoke bin/k8s-template --mappings=tests/mappings.yaml --template=tests/template.yaml
	bin/k8s-template --inplace --template=tests/env.yaml
	bin/k8s-template --mappings=tests/mappings.yaml --template=tests/env.yaml
	bin/k8s-template --mappings=tests/empty.yaml --template=tests/env.yaml
