SHELL=/bin/bash

libdep="gopkg.in/yaml.v2" 

all: init get save build test
	@echo make targets init to initialize godeps, get, save, test and build

init: get save

build: 
	godep go build -a -ldflags '-s'
get:
	godep get $(libdep)
save:
	godep save
test: 
	godep go test -v

