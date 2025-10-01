package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/cli/internal/webhook"
	"github.com/amp-labs/cli/logger"
	"github.com/spf13/cobra"
)

var (
	ErrInvalidEventFormat = errors.New("invalid event format, expected 'provider.event'")
	fixtureFile           string
	rawJSON               string
	interactive           bool
	listenPort            string
	triggerCommand        = &cobra.Command{
		Use:   "trigger [provider.event]",
		Short: "Trigger a webhook event",
		Long: `Trigger a webhook event to be sent to the local listener.
This command sends a webhook event to the local listener using fixture data.
You can specify a built-in event or provide your own JSON payload.

Examples:
  amp trigger stripe.payment_intent.created
  amp trigger hubspot.contact.created --interactive
  amp trigger stripe.payment_intent.created --fixture ./my-custom-event.json
  amp trigger custom.event --raw '{"key": "value"}'`,
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE:   runTrigger,
	}
)

func init() {
	triggerCommand.Flags().StringVar(&fixtureFile, "fixture", "", "Path to a custom fixture file")
	triggerCommand.Flags().StringVar(&rawJSON, "raw", "", "Raw JSON payload to send")
	triggerCommand.Flags().BoolVar(&interactive, "interactive", false, "Open editor before sending")
	triggerCommand.Flags().StringVar(&listenPort, "port", "", "Port of the local listener (default: auto-detect)")
	rootCmd.AddCommand(triggerCommand)
}

func runTrigger(cmd *cobra.Command, args []string) error {
	eventName := args[0]
	provider, event := webhook.ParseEvent(eventName)

	if event == "" {
		return fmt.Errorf("%w: %q", ErrInvalidEventFormat, eventName)
	}

	// Determine which payload to use
	var payload []byte

	var err error

	switch {
	case rawJSON != "":
		// Use raw JSON provided via command line
		payload = []byte(rawJSON)

		// Validate it's valid JSON
		var jsonObj interface{}
		if err := json.Unmarshal(payload, &jsonObj); err != nil {
			return fmt.Errorf("invalid JSON provided: %w", err)
		}
	case fixtureFile != "":
		// Use custom fixture file
		payload, err = webhook.LoadFixture(provider, event, fixtureFile)
		if err != nil {
			return err
		}
	default:
		// Use built-in fixture
		payload, err = webhook.LoadFixture(provider, event, "")
		if err != nil {
			return fmt.Errorf("no fixture found for %s.%s: %w", provider, event, err)
		}
	}

	// Let user edit the payload if requested
	if interactive {
		payload, err = openInEditor(payload)
		if err != nil {
			return fmt.Errorf("failed to edit payload: %w", err)
		}
	}

	// Send the webhook
	fmt.Fprint(os.Stdout, "🚀 Triggering webhook: "+eventName+"\n")

	return sendWebhook(payload)
}

// openInEditor opens the JSON payload in the default editor.
func openInEditor(data []byte) ([]byte, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "amp-webhook-*.json")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	// Write the data to the file
	if _, err := tmpFile.Write(data); err != nil {
		return nil, err
	}

	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	// Get the editor command
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}

	// Open the editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Read the updated content
	return os.ReadFile(tmpFile.Name())
}

func sendWebhook(payload []byte) error {
	port := getListenerPort()

	url := "http://127.0.0.1:" + port

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	const clientTimeout = 5 * time.Second
	client := &http.Client{Timeout: clientTimeout}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}

	defer resp.Body.Close()

	fmt.Fprint(os.Stdout, "✅ Sent webhook → "+strconv.Itoa(resp.StatusCode)+" "+resp.Status+"\n")

	return nil
}

func getListenerPort() string {
	if listenPort != "" {
		return listenPort
	}

	// Try to get the port from the port file
	dir, err := os.UserCacheDir()
	if err != nil {
		logger.Debugf("error getting user cache dir: %v", err)

		return "4242", nil // Default fallback port
	}

	portFile := filepath.Join(dir, "ampersand", "webhook-port")

	data, err := os.ReadFile(portFile)
	if err != nil {
		logger.Debug("could not find webhook port file, using default port 4242")

		return "4242" // Default fallback port
	}

	return strings.TrimSpace(string(data))
}
