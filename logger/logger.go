package logger

import (
	"fmt"
	"github.com/wakeful-deployment/operator/global"
)

func Info(content string) {
	Log("INFO: ", content)
}

func Error(content string) {
	Log("ERROR: ", content)
}

func Log(prefix string, content string) {
	if global.Config.Verbose {
		fmt.Println(fmt.Sprintf("%s%s", prefix, content))
	}
}
