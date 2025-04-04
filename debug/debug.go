package debug

import (
	"log"
	"os"
)

var enabled bool

// Enable turns on debug logging to stderr.
func Enable() {
	enabled = true
	log.SetOutput(os.Stderr)
}

// Disable turns off debug logging.
func Disable() {
	enabled = false
}

// Printf logs the message if debugging is enabled.
func Printf(format string, v ...interface{}) {
	if enabled {
		log.Printf(format, v...)
	}
}
