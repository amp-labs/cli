package cmd

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const executeUnknownCommandEnv = "AMP_TEST_EXECUTE_UNKNOWN_COMMAND"

func TestExecuteExitsWithFailureForUnknownCommand(t *testing.T) {
	t.Parallel()

	if os.Getenv(executeUnknownCommandEnv) == "1" {
		os.Args = []string{"amp", "definitely-not-a-command"}

		Execute()

		return
	}

	// This test intentionally executes the current test binary.
	//nolint:gosec
	command := exec.CommandContext(
		t.Context(), os.Args[0], "-test.run=^TestExecuteExitsWithFailureForUnknownCommand$",
	)

	command.Env = append(os.Environ(), executeUnknownCommandEnv+"=1")

	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("Execute() exited successfully for an unknown command: %s", output)
	}

	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("Execute() returned %T, want an exit error", err)
	}

	if exitError.ExitCode() != 1 {
		t.Fatalf("Execute() returned exit code %d, want 1", exitError.ExitCode())
	}

	if !strings.Contains(string(output), `unknown command "definitely-not-a-command"`) {
		t.Fatalf("Execute() output = %q, want unknown command error", output)
	}
}
