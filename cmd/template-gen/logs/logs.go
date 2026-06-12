package logs

import (
	"log"
)

var debugLogs bool

func Init(debug bool) {
	debugLogs = debug
}

func Debug(v ...any) {
	if debugLogs {
		log.Println(v...)
	}
}
