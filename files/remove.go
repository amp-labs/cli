package files

import (
	"os"

	"github.com/amp-labs/cli/logger"
)

func Remove(filepath string) {
	err := os.Remove(filepath)
	if err != nil {
		logger.FatalErr("Failed to remove folder", err)
		return
	}
}
