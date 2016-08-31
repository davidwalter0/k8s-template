package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Logger *log.Logger

var (
	Trace   *log.Logger
	Info    *log.Logger
	Debug   *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Plain   *log.Logger
)

func Init(
	traceHandle io.Writer,
	debugHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
	plainHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(debugHandle,
		"DEBUG: ",
		log.Llongfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Plain = log.New(plainHandle, "", 0)
}

func Args2String(args []interface{}) string {
	return fmt.Sprintf("%v", args...)
}

func init() {
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stdout, os.Stderr, os.Stdout)
}
