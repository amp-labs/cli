package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/cli/flags"
)

func Info(a ...any) {
	fmt.Fprint(os.Stdout, fmt.Sprintln(a...))
}

func Infof(msg string, a ...any) {
	Info(fmt.Sprintf(msg, a...))
}

func Debug(msg string) {
	if flags.GetDebugMode() {
		fmt.Fprint(os.Stdout, time.Now().Format(time.RFC3339)+" DEBUG: "+msg+"\n")
	}
}

func Debugf(msg string, a ...any) {
	Debug(fmt.Sprintf(msg, a...))
}

func Fatal(msg string) {
	Info(msg)
	os.Exit(1)
}

func FatalErr(msg string, err error) {
	Fatal(fmt.Sprintf("%v\nerror: %v", msg, err))
	PrintDebugTip()
}

func PrintDebugTip() {
	fmt.Fprint(os.Stdout, "For more information, run again with --debug\n")
}
