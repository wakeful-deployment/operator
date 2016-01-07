package logger

import (
	"fmt"
)

var Verbose bool

func Info(content string) {
	Log("INFO: ", content)
}

func Error(content string) {
	Log("ERROR: ", content)
}

func Log(prefix string, content string) {
	if Verbose {
		fmt.Println(fmt.Sprintf("%s%s", prefix, content))
	}
}
