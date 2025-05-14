package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LoadFixture loads a fixture file and replaces template placeholders
func LoadFixture(provider, event, customPath string) ([]byte, error) {
	var path string
	if customPath != "" {
		path = customPath
	} else {
		// Default path is in the internal fixtures directory
		path = filepath.Join("internal", "fixtures", provider, event+".json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file: %w", err)
	}

	// Replace template tokens
	now := time.Now().UTC().Format(time.RFC3339)
	data = bytes.ReplaceAll(data, []byte("{{NOW}}"), []byte(now))

	// Validate it's valid JSON
	var jsonObj interface{}
	if err := json.Unmarshal(data, &jsonObj); err != nil {
		return nil, fmt.Errorf("fixture contains invalid JSON: %w", err)
	}

	return data, nil
}

// ParseEvent parses an event string into provider and event name
// Format: provider.event_name (e.g., "stripe.payment_intent.created")
func ParseEvent(event string) (provider, eventName string) {
	parts := strings.SplitN(event, ".", 2)
	if len(parts) < 2 {
		return event, ""
	}
	return parts[0], parts[1]
}

// PrettyPrintJSON formats and prints JSON data to stdout with colors
func PrettyPrintJSON(data []byte) error {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("\n→ Received webhook event:\n%s\n", prettyJSON.String())
	fmt.Println("------------------------------------------")
	return nil
} 