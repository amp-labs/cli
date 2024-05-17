package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/cli/flags"
)

func Info(a ...any) {
	fmt.Println(a...)
}

func Infof(msg string, a ...any) {
	Info(fmt.Sprintf(msg, a...))
}

func Debug(msg string) {
	if flags.GetDebugMode() {
		fmt.Println(time.Now().Format(time.RFC3339), "DEBUG:", msg)
	}
}

func Debugf(msg string, a ...any) {
	Debug(fmt.Sprintf(msg, a...))
}

func Fatal(msg string) {
	Infof(msg)
	os.Exit(1)
}

func FatalErr(msg string, err error) {
	Fatal(fmt.Sprintf("%v\nerr: %v", msg, err))
	PrintDebugTip()
}

func PrintDebugTip() {
	fmt.Println("For more information, run with --debug")
}
