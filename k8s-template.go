/*
template:

Accepts a yaml formatted array of maps of mappings for substitution name:

name: the name used for pattern matching by golang template replacement
value:
optional: bool flags:
file:[true|false] -- read from the file named in value
base64:[true|false] -- convert the final text to base64
env:[true|false] -- read from the environment

Replace golang template formatted targets with values specified in the
mappings names.

Formatted --template input file has template replacement performed on it. The
resulting replacement target is printed to stdout.

The yaml formattted mappings file can be similar to the following:

- name: PrivateKey
  base64: false
  file: true
  value: ~/.ssh/id_rsa

- name: PublicKey
  base64: false
  value: ~/.ssh/id_rsa.pub

A substitution of the corresponding golang template can specify base64
translation later

*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/davidwalter0/logger"
	"github.com/davidwalter0/transform"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"text/template"
)

var TemplateFile = flag.String("template", "", "file with templates to replace")
var MappingsFile = flag.String("mappings", "mappings.yaml", "describe the replacement values")
var version = flag.Bool("version", false, "print build and git commit as a version string")
var debug = flag.Bool("debug", false, "dump additional debugging information on template apply failure")

var TemplateText []byte
var ReplacementText []byte

var Info = logger.Info
var Elog = logger.Error
var Plain = logger.Plain

var Build string  // from the build ldflag options
var Commit string // from the build ldflag options

/*
TemplateMapping

If file: text, text names a file. Use the content of the named file as the value
If env: text use text as the env var as the source value
If value: text use text as the source value

*/

type TemplateMapping struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Base64 bool   `json:"base64,omitempty"`
	File   bool   `json:"file,omitempty"`
	Env    bool   `json:"env,omitempty"`
}

// Base64EncodeString transform input to base64 encoded data
func Base64Encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

var fmap = template.FuncMap{
	"base64Encode": Base64Encode,
}

// var debugFile *os.File = os.Stdout
var debugText string

func init() {
	flag.Parse()
	if *Version {
		array := strings.Split(os.Args[0], "/")
		me := array[len(array)-1]
		fmt.Println(me, "Build:", Build, "Commit:", Commit)
	}
	// if len(*debugFileName) > 0 {
	// 	var err error
	// 	debugFile, err = os.Create(*debugFileName)
	// 	debugFile, err = os.OpenFile(*debugFileName, os.O_RDWR, 0666)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err)
	// 		os.Exit(3)
	// 	}
	// }
}

func Usage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(3)
}

func Trace() {
	pc := make([]uintptr, 10)
	runtime.Callers(10, pc)
	for i := 0; i < 10; i++ {
		if pc[i] == 0 {
			break
		}
		f := runtime.FuncForPC(pc[i])
		file, line := f.FileLine(pc[i])
		Info.Printf("%s:%d %s\n", file, line, f.Name())
	}
}

func RecoverWithMessage(step string, exitOnException bool, failureExitCode int) {
	if r := recover(); r != nil {
		Info.Printf("Recovered step[%s] with info %v\n", step, r)
		Trace()
		pc := make([]uintptr, 10)
		runtime.Callers(5, pc)
		f := runtime.FuncForPC(pc[1])
		file, line := f.FileLine(pc[1])
		Info.Printf("call failed at or near %s:%d %s\n", file, line, f.Name())
		if exitOnException {
			os.Exit(failureExitCode)
		}
	}
}

func (tm *TemplateMapping) Parse(InData map[string]interface{}) {
	defer RecoverWithMessage("Parse", false, 3)
	for key, value := range InData {
		switch key {
		case "name":
			tm.Name = value.(string)
		case "value":
			tm.Value = value.(string)
		case "base64":
			tm.Base64 = value.(bool)
		case "file":
			tm.File = value.(bool)
		case "env":
			tm.Env = value.(bool)
		}
	}

	if tm.Env {
		tm.Value = os.Getenv(tm.Value)
	}

	if tm.File {
		path := tm.Value
		if len(path) > 2 && path[:2] == "~/" {
			path = strings.Replace(path, "~/", os.Getenv("HOME")+"/", 1)
		}
		tm.Value = string(Load(path))
	}

	if tm.Base64 {
		tm.Value = Base64Encode(tm.Value)
	}

	if *debug {
		debugText += fmt.Sprintf("name: %s len(value): %d base64: %v file: %v env: %v\n",
			tm.Name, len(tm.Value), tm.Base64, tm.File, tm.Env)
	}
}

func Load(filename string) []byte {
	var err error
	var text []byte

	if len(filename) > 0 {
		text, err = ioutil.ReadFile(filename)
		if err != nil {
			Elog.Printf("%v.\n", err)
			os.Exit(3)
		}
	} else {
		Usage()
	}
	return text
}

type ReplacementMapping map[string]string

var MappingDefinition []map[string]interface{}

var Mapping ReplacementMapping = make(ReplacementMapping)

func main() {
	defer RecoverWithMessage("main", false, 3)
	var err error
	TemplateText = Load(*TemplateFile)
	ReplacementText = Load(*MappingsFile)
	data, err := transform.Yaml2Json(ReplacementText)
	if err != nil {
		fmt.Println(err, "error transforming Yaml2Json")
		os.Exit(3)
	}
	_ = json.Unmarshal(data, &MappingDefinition)
	for _, InData := range MappingDefinition {
		var tm TemplateMapping
		tm.Parse(InData)
		Mapping[tm.Name] = tm.Value
	}
	TemplateApply(Mapping, TemplateText)
}

func TemplateApply(mapping ReplacementMapping, InText []byte) { // string {
	defer RecoverWithMessage("TemplateApply", false, 3)
	tmpl, err := template.New(*TemplateFile).Funcs(fmap).Parse(string(InText))
	if err != nil {
		Elog.Fatal(err)
	}
	err = tmpl.Execute(os.Stdout, mapping)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintf(os.Stderr, "Is the template file missing a mapping?\nCheck the line number of the error to see if the mapping file has that argument.")
		if len(debugText) > 0 {
			fmt.Fprintf(os.Stderr, "\nInput debug mappings:\n")
			fmt.Fprintln(os.Stderr, debugText)
		}
		os.Exit(3)
	}
	os.Stdout.Sync()
}
