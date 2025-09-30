package logger

import (
	"fmt"
	"os"
)

var enableDebugLogging bool = false
var envbleWarningLogging bool = false

func Debug(format string, a ...any) {
	if !enableDebugLogging {
		return
	}
	fmt.Fprintf(os.Stderr, fmt.Sprintf("\033[0;36mdebug:\033[0m %s\r\n", format), a...)
}

func Warning(format string, a ...any) {
	if !envbleWarningLogging {
		return
	}
	fmt.Fprintf(os.Stderr, fmt.Sprintf("\033[0;33mWarning: %s\033[0m\r\n", format), a...)
}
