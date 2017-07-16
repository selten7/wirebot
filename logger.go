package wirebot

import (
	"io/ioutil"
	"log"
)

var (
	discardLogger Logger = log.New(ioutil.Discard, "", 0)
)

// Logger is a basic logging interface.
type Logger interface {
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
