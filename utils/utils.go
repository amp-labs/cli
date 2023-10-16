package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/amp-labs/cli/logger"
)

func GetWorkingDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		logger.FatalErr("Unable to get working directory", err)

		return ""
	}

	return workingDir
}

// NewTimestampedZipName returns a zip name with a timestamp.
func NewTimestampedZipName() string {
	return fmt.Sprintf("amp_%d.zip", time.Now().Unix())
}
