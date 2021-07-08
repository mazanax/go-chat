package logger

import (
	"fmt"
	"io"
	"os"
)

var debugOutput io.Writer
var errorOutput io.Writer

func SetDebugOutputFile(file io.Writer) {
	debugOutput = file
}

func SetErrorOutputFile(file io.Writer) {
	errorOutput = file
}

func Debug(message string, arguments ...interface{}) {
	if debugOutput == nil {
		debugOutput = os.Stdout
	}

	_, _ = fmt.Fprintf(debugOutput, message, arguments...)
}

func Error(message string, arguments ...interface{}) {
	if errorOutput == nil {
		errorOutput = os.Stderr
	}

	_, _ = fmt.Fprintf(errorOutput, message, arguments...)
}

func Fatal(message string, arguments ...interface{}) {
	if errorOutput != nil {
		errorOutput = os.Stderr
	}

	_, _ = fmt.Fprintf(errorOutput, message, arguments...)
	os.Exit(1)
}
