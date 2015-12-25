/*
template:

Accepts a yaml formatted array of maps of mappings for substitution name:

name: the name used for pattern matching by golang template replacement

value: the text value to replace the mapping with

optional: flags:

file:  [true|false] -- read from the file named in value
base64:[true|false] -- convert the final text to base64
env:   [true|false] -- read from the environment

Replace golang template formatted targets with values specified in the
mappings names.

Formatted --template input file has template replacement performed on it. The
resulting replacement target is printed to stdout.

--preprocess can be used to perform self referential mappings

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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/davidwalter0/logger"
	"github.com/davidwalter0/transform"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var TemplateFile = flag.String("template", "", "file with templates to replace, or if not set, act as a filter.")
var MappingsFile = flag.String("mappings", "", "describe the replacement values")
var preprocess = flag.Bool("preprocess", false, "dump to standard output the preprocessed, template replacements of a mapping skipping file inclusion")
var version = flag.Bool("version", false, "print build and git commit as a version string")
var debug = flag.Bool("debug", false, "dump additional debugging information on template apply failure")

var TemplateText []byte
var ReplacementMappingSourceText []byte

var Info = logger.Info
var Elog = logger.Error
var Plain = logger.Plain

var Build string  // from the build ldflag options
var Commit string // from the build ldflag options

var templateRegex *regexp.Regexp

/*
TemplateMapping

If file: text, text names a file. Use the content of the named file as the value
If env: text use text as the env var as the source value
If value: text use text as the source value

*/

type ReplacementMapping map[string]string
type FileMapped map[string]bool
type Base64Mapped map[string]bool

var MappingDefinition []map[string]interface{}
var Mapping ReplacementMapping = make(ReplacementMapping)
var fileMapped = make(FileMapped)
var base64Mapped = make(Base64Mapped)

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

// Split a string to an array of strings on a space character
func Split(text string) []string {
	return strings.Split(Trim(text), " ")
}

// First item in an array split on spaces, using Split
func First(text string) string {
	array := Split(text)

	if len(array) > 0 {
		text = array[0]
	} else {
		text = ""
	}
	return text
}

// Nth zero offset item in the array after the text argument is split on spaces
func Nth(nstr string, text string) string {
	n, _ := strconv.Atoi(nstr)
	array := Split(text)
	// if *debug {
	// 	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// 	if err != nil {
	// 		Elog.Fatalf("error opening file: %v", err)
	// 	}
	// 	defer f.Close()

	// 	Plain.SetOutput(f)
	// 	Plain.Printf("ns %s text %s\n", nstr, text)
	// 	Plain.Printf("nstr %s n %d text %s\n", nstr, n, text)
	// 	Plain.Printf("array %v\n", array)
	// 	Plain.Printf("len(array) = %d array[%d] array[n]  %v\n", len(array), n, array[n])
	// }
	if len(array) > n {
		return array[n]
	} else {
		return ""
	}
}

// Trim spaces from a string
func Trim(text string) string {
	return strings.Trim(text, " ")
}

// Delimit a space separated string with delimiter [ default comma ',' ]
func Delimit(text string, delimiter string) (o string) {
	if len(delimiter) == 0 {
		delimiter = ","
	}
	array := Split(text)
	for i, x := range array {
		if i > 0 {
			o += delimiter
		}
		o += x
	}
	return o
}

// zip 2 space separated lists with a separator char like "."
// "a b c" "1 2 3" "." -> "a.1 a.2 a.3 b.1 b.2 b.3"
// split list1 list2 and append with separator
func Zip(list1, list2, separator string) string {
	l1 := strings.Split(Trim(list1), " ")
	l2 := strings.Split(Trim(list2), " ")
	if len(separator) == 0 {
		separator = "-"
	}
	text := ""
	for _, x := range l2 {
		for j, y := range l1 {
			if j < len(l1) {
				text += " "
			}
			text += x + separator + y
		}
	}
	return text
}

// return the text of the index of search [find] from in text
func Index(find, in string) (text string) {
	array := strings.Split(Trim(in), " ")
	for i, x := range array {
		if find == x {
			text = strconv.Itoa(i)
			break
		}
	}
	return text
}

func ZipPrefix(text, prefix, separator string) []string {
	if len(separator) == 0 {
		separator = "-"
	}
	array := Split(text)
	text = ""
	for i, x := range array {
		array[i] = prefix + separator + x
	}
	return array
}

func ZipSuffix(text, suffix, separator string) []string {
	if len(separator) == 0 {
		separator = "-"
	}
	array := Split(text)
	text = ""
	for i, x := range array {
		array[i] = x + separator + suffix
	}
	return array
}

func Cat(in ...string) string {
	text := ""
	for _, x := range in {
		text += text + x
	}
	return text
}

// func Switch(xin ...string) string {
// 	text := ""
// 	for _, x := range in {
// 		text += text + x
// 	}
// 	return text
// }

var fmap = template.FuncMap{
	"cat":          Cat,
	"nth":          Nth,
	"delimit":      Delimit, // replace space with ,
	"base64Encode": Base64Encode,
	"split":        Split,
	"zip":          Zip,
	"zipPrefix":    ZipPrefix,
	"zipSuffix":    ZipSuffix,
	"trim":         Trim,
	"first":        First,
	"index":        Index,
}

// var debugFile *os.File = os.Stdout
var debugText string

func init() {
	flag.Parse()
	if *version {
		array := strings.Split(os.Args[0], "/")
		me := array[len(array)-1]
		fmt.Println(me, "Build:", Build, "Commit:", Commit)
	}

	templateRegex, _ = regexp.Compile("{{.*}}")
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
		if !templateRegex.MatchString(tm.Value) && !*preprocess {
			path := tm.Value
			if len(path) > 2 && path[:2] == "~/" {
				path = strings.Replace(path, "~/", os.Getenv("HOME")+"/", 1)
			}
			tm.Value = string(Load(path))
		}

		if *preprocess {
			fileMapped[tm.Name] = true
		}
	}

	if tm.Base64 {
		if *preprocess {
			base64Mapped[tm.Name] = true
		} else {
			tm.Value = Base64Encode(tm.Value)
		}
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
		text, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			Elog.Printf("%v.\n", err)
			os.Exit(3)
		}
		// Usage()
	}
	return text
}

var IOStdin bool = false

func main() {
	defer RecoverWithMessage("main", false, 3)
	var err error

	if len(*TemplateFile) == 0 && len(*MappingsFile) == 0 {
		IOStdin = true
		TemplateText, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			Elog.Printf("%v.\n", err)
			os.Exit(3)
		}
		ReplacementMappingSourceText = TemplateText
	} else {
		TemplateText = Load(*TemplateFile)
		if len(*TemplateFile) == 0 {
			var stdin = "stdin"
			TemplateFile = &stdin
		}
		ReplacementMappingSourceText = Load(*MappingsFile)
	}

	data, err := transform.Yaml2Json(ReplacementMappingSourceText)
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

	if *debug {
		f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Elog.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		Plain.SetOutput(f)
	}
	SelfReference(&Mapping)
	if *preprocess {
		Preprocess(Mapping, IOStdin)
	}
	if !IOStdin && !*preprocess {
		TemplateApply(Mapping, TemplateText)
	}
}

// Apply template reconciliation to mappings templates to interpolate
// template local self referential mappings. After this is done, the
// local template references should have been replaced.
func SelfReference(m *ReplacementMapping) {
	Mapping := *m
	var done bool = false
	// Continue while there are replacements being made
	for !done {
		done = true
		for k, v := range Mapping {
			before := v
			for templateRegex.MatchString(before) {
				out := TemplateApplyString(Mapping, v)
				after := out
				// If there is a mapping without changes, this has been
				// processed as much as it can be for now.
				if before == after {
					break
				} else {
					done = false
				}
				Mapping[k] = after
				before = after
			}
		}
	}
}

func Preprocess(Mapping ReplacementMapping, dump bool) {
	var keys []string
	var OutMap []TemplateMapping = make([]TemplateMapping, 0)
	for k, _ := range Mapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		var T TemplateMapping
		T.Name = key
		T.Value = Mapping[key]
		T.File = false
		if fileMapped[key] {
			T.File = true
		}
		if base64Mapped[key] {
			T.Base64 = true
		}

		OutMap = append(OutMap, T)
	}
	if dump {
		fmt.Println(Json2Yaml([]byte(Jsonify(OutMap))))
	}
}

func Jsonify(data interface{}) string {
	var err error
	data, err = transform.TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	s, err := json.MarshalIndent(data, "", "  ") // spaces)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}

func Json2Yaml(input []byte) string {
	var data interface{}
	var err = json.Unmarshal(input, &data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	data, err = transform.TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	output, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(output)
}

func Yamlify(data interface{}) string {
	data, err := transform.TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	s, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}

func TemplateApplyString(mapping ReplacementMapping, text string) string { // string {
	defer RecoverWithMessage("TemplateApplyString", false, 3)
	buffer := new(bytes.Buffer)
	tmpl, err := template.New("TemplateApplyString").Funcs(fmap).Parse(text)
	err = tmpl.Execute(buffer, mapping)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintf(os.Stderr, "Is the template file missing a mapping?\nCheck the line number of the error to see if the mapping file has that argument.")
		if len(debugText) > 0 {
			fmt.Fprintf(os.Stderr, "\nInput debug mappings:\n")
			fmt.Fprintln(os.Stderr, debugText)
		}
	}
	return buffer.String()
}

func TemplateApply(mapping ReplacementMapping, ttext []byte) { // string {
	defer RecoverWithMessage("TemplateApply", false, 3)
	text := string(ttext)
	for templateRegex.MatchString(text) {
		after := TemplateApplyString(Mapping, text)
		// If there is a mapping without changes, this has been
		// processed as much as it can be for now.
		if text == after {
			break
		}
		text = after
	}
	fmt.Fprintln(os.Stdout, text)
	os.Stdout.Sync()
}
