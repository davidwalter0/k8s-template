Initially based on package yaml2json with an MIT license agreement. [![Build Status](https://travis-ci.org/davidwalter0/transform.svg?branch=master)](https://travis-ci.org/davidwalter0/transform)

bronze1man/yaml2json

The default makefile is configured to use godeps for vendoring.

```
make -k get save build test
GO15VENDOREXPERIMENT=1 GOPATH=/go /go/bin/godep get "gopkg.in/yaml.v2" 
GO15VENDOREXPERIMENT=1 GOPATH=/go /go/bin/godep save
GO15VENDOREXPERIMENT=1 GOPATH=/go /go/bin/godep go build -a -ldflags '-s'
GO15VENDOREXPERIMENT=1 GOPATH=/go /go/bin/godep go test -v
=== RUN   TestUnhtmlize
--- PASS: TestUnhtmlize (0.00s)
=== RUN   TestJson2Yaml
--- PASS: TestJson2Yaml (0.00s)
=== RUN   TestYaml2Json
--- PASS: TestYaml2Json (0.00s)
=== RUN   TestYaml2JsonIndent
--- PASS: TestYaml2JsonIndent (0.00s)
PASS
ok  	github.com/davidwalter0/transform	0.003s
```
