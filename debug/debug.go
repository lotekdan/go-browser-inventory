package debug

import (
	"log"
)

var debugEnabled = false

func SetEnabled(enabled bool) {
	debugEnabled = enabled
}

func Printf(format string, v ...interface{}) {
	if debugEnabled {
		log.Printf(format, v...)
	}
}
