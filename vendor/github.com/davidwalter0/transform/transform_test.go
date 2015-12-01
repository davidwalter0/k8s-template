package transform

/*
This does not do formal validation on a language specification.
These are text comparisons of transformations to/from json/yaml
*/
import (
	"testing"
)

func TestUnhtmlize(t *testing.T) {
	var expect string = "<html>\n<body>Hello World! } &;\n</body>\n</html>\n"
	var text string = "\u003chtml\u003e\n<body>Hello World! } \u0026;\n</body>\n</html>\n"
	var bytes []byte = []byte(text)
	lhs := Unhtmlize(bytes)
	if string(lhs) != expect {
		t.Errorf("Expected [%v] \n`%s`\n got \n`%s`\n", bytes, expect, text)
	}
}

func TestJson2Yaml(t *testing.T) {
	var json string = `{
  "apiVersion": "v0.1",
  "kind": "test",
  "spec": [
    {
      "name": "test-cmd",
      "protocol": "tcp",
      "request": {
        "spec": {
          "arg": [],
          "env": {},
          "url": "uri"
        }
      },
      "response": {
        "ack": {},
        "nak": {},
        "x": true
      }
    }
  ]
}`

	var yaml string = `apiVersion: v0.1
kind: test
spec:
- name: test-cmd
  protocol: tcp
  request:
    spec:
      arg: []
      env: {}
      url: uri
  response:
    ack: {}
    nak: {}
    x: true
`
	got, err := Json2Yaml([]byte(json))
	if string(yaml) != string(got) {
		t.Errorf("Expected\n`%s`\n got \n`%s`\n error %v\n", json, got, err)
	}
}

func TestYaml2Json(t *testing.T) {
	var json string = `{"apiVersion":"v0.1","kind":"test","spec":[{"name":"test-cmd","protocol":"tcp","request":{"spec":{"arg":[],"env":{},"url":"uri"}},"response":{"ack":{},"nak":{},"x":true}}]}`
	var yaml string = `apiVersion: v0.1
kind: test
spec:
- name: test-cmd
  protocol: tcp
  request:
    spec:
      arg: []
      env: {}
      url: uri
  response:
    ack: {}
    nak: {}
    x: true
`
	got, err := Yaml2Json([]byte(yaml))
	if string(json) != string(got) {
		t.Errorf("Expected\n`%s`\n got \n`%s`\n error %v\n", json, got, err)
	}
}

func TestYaml2JsonIndent(t *testing.T) {
	var json string = `{
  "apiVersion": "v0.1",
  "kind": "test",
  "spec": [
    {
      "name": "test-cmd",
      "protocol": "tcp",
      "request": {
        "spec": {
          "arg": [],
          "env": {},
          "url": "uri"
        }
      },
      "response": {
        "ack": {},
        "nak": {},
        "x": true
      }
    }
  ]
}`

	var yaml string = `apiVersion: v0.1
kind: test
spec:
- name: test-cmd
  protocol: tcp
  request:
    spec:
      arg: []
      env: {}
      url: uri
  response:
    ack: {}
    nak: {}
    x: true
`
	got, err := Yaml2JsonIndent([]byte(yaml))
	if string(json) != string(got) {
		t.Errorf("Expected\n`%s`\n got \n`%s`\n error %v\n", json, got, err)
	}
}
