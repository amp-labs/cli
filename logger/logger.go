package logger

import (
	"fmt"
	"os"
)

func Info(msg string) {
	fmt.Println(msg)
}

func Infof(msg string, a ...any) {
	fmt.Println(fmt.Sprintf(msg, a...))
}

func Debug(msg string) {
	// TODO: only print if debug mode is on
	// TODO: print other metadata that might be helpful for debugging
	fmt.Println("DEBUG:", msg)
}

func Debugf(msg string, a ...any) {
	Debug(fmt.Sprintf(msg, a...))
}

func Fatal(msg string) {
	Debug(msg)
	os.Exit(1)
}

func FatalErr(msg string, err error) {
	Fatal(fmt.Sprintf("%v, err: %v", msg, err))
}
