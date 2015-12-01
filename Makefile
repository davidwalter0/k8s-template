.PHONY: deps
# GOPATH=/go
SHELL=/bin/bash
MAKEFILE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CURRENT_DIR := $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))
DIR=$(MAKEFILE_DIR)

targets: bin/k8s-template

build: targets

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

	if go version|grep -q 1.4 ; then											\
	    args="-s -w -X main.Build $$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit $$(git log --format=%hash-%aI -n1)";	\
	fi;															\
	if go version|grep -q 1.5 ; then											\
	    args="-s -w -X main.Build=$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit=$$(git log --format=%hash-%aI -n1)";	\
	fi;															\
	CGO_ENABLED=0 go get --tags netgo -ldflags "$${args}" ;									\
	CGO_ENABLED=0 go build --tags netgo -ldflags "$${args}" -o $@ $^ ;


init: get save

get: 
	GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH} $(GOPATH)/bin/godep get $(libdep)
save:
	GO15VENDOREXPERIMENT=1 GOPATH=${GOPATH} $(GOPATH)/bin/godep save
clean:
	@echo cleaning up temporary files
	@echo rm -f $(targets)
	@rm -f $(targets) .dep $(package_dir)/.dep

