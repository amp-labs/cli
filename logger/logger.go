package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/cli/flags"
)

func Info(a ...any) {
	fmt.Fprintln(os.Stdout, a...)
}

func Infof(msg string, a ...any) {
	Info(fmt.Sprintf(msg, a...))
}

// Warn prints a warning to stderr, so that it never mixes into output a caller may be
// piping or parsing (e.g. 'amp get:region', or a list command with --format json).
func Warn(msg string) {
	fmt.Fprintf(os.Stderr, "Warning: %s\n", msg)
}

func Warnf(msg string, a ...any) {
	Warn(fmt.Sprintf(msg, a...))
}

func Debug(msg string) {
	if flags.GetDebugMode() {
		fmt.Fprintf(os.Stdout, "%s DEBUG: %s\n", time.Now().Format(time.RFC3339), msg)
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
