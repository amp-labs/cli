package flags

import (
	"testing"

	"github.com/amp-labs/cli/utils"
	"github.com/spf13/cobra"
)

func TestGetOutputFormatForCommandUsesOwningFlag(t *testing.T) {
	t.Parallel()

	jsonCommand := &cobra.Command{Use: "json"}
	jsonCommand.Flags().String("format", "json", "")

	yamlCommand := &cobra.Command{Use: "yaml"}
	yamlCommand.Flags().String("format", "json", "")

	err := yamlCommand.Flags().Set("format", "yaml")
	if err != nil {
		t.Fatalf("set format: %v", err)
	}

	if got := GetOutputFormatForCommand(jsonCommand); got != utils.JSON {
		t.Fatalf("JSON command format = %q, want json", got)
	}

	if got := GetOutputFormatForCommand(yamlCommand); got != utils.YAML {
		t.Fatalf("YAML command format = %q, want yaml", got)
	}
}
