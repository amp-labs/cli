package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	destinations []string
	skipConfirm  bool
	ngrokCommand = &cobra.Command{
		Use:    "ngrok",
		Short:  "Ngrok tunnel management",
		Long:   "Configure Ampersand destinations to use ngrok tunnels for local development.",
		Hidden: true,
		RunE:   runNgrok,
	}
)

type Destination struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (d Destination) String() string {
	if d.Name != "" {
		return fmt.Sprintf("%s (%s)", d.Name, d.Id)
	}

	return d.Id
}

type ngrokTunnel struct {
	PublicURL string `json:"public_url"`
}

type ngrokResponse struct {
	Tunnels []ngrokTunnel `json:"tunnels"`
}

func init() {
	ngrokCommand.Flags().StringSliceVarP(&destinations,
		"destination", "D", []string{},
		"Destination names or IDs (UUIDs)")

	ngrokCommand.Flags().BoolVarP(&skipConfirm,
		"yes", "y", false,
		"Skip confirmation prompts")

	rootCmd.AddCommand(ngrokCommand)
}

func getPublicNgrokUrl(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:4040/api/tunnels", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create ngrok API request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to ngrok API: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error closing response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ngrok API returned status %d", resp.StatusCode)
	}

	var ngrokResp ngrokResponse
	if err := json.NewDecoder(resp.Body).Decode(&ngrokResp); err != nil {
		return "", fmt.Errorf("failed to parse ngrok response: %w", err)
	}

	return chooseNgrokTunnel(&ngrokResp)
}

func chooseNgrokTunnel(ngrokResp *ngrokResponse) (string, error) {
	if len(ngrokResp.Tunnels) == 0 {
		return "", fmt.Errorf("no ngrok tunnels found")
	}

	if len(ngrokResp.Tunnels) == 1 {
		return ngrokResp.Tunnels[0].PublicURL, nil
	}

	urls := make([]string, len(ngrokResp.Tunnels))
	for i, tunnel := range ngrokResp.Tunnels {
		urls[i] = tunnel.PublicURL
	}

	prompt := promptui.Select{
		Label:  "Choose ngrok tunnel",
		Items:  urls,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select tunnel: %w", err)
	}

	return urls[idx], nil
}

func waitForNgrok(ctx context.Context) error {
	const (
		maxDuration   = 5 * time.Minute
		retryInterval = 2 * time.Second
		address       = "localhost:4040"
	)

	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err == nil {
		conn.Close()
		return nil
	}

	deadline := time.Now().Add(maxDuration)
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	needsNewline := false

	defer func() {
		if needsNewline {
			_, _ = os.Stdout.Write([]byte("\n"))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for ngrok on %s after %v", address, maxDuration)
			}

			conn, err := net.DialTimeout("tcp", address, time.Second)
			if err == nil {
				conn.Close()
				return nil
			}

			_, _ = os.Stdout.Write([]byte("."))
			_ = os.Stdout.Sync()
			needsNewline = true
		}
	}
}

func runNgrok(cmd *cobra.Command, _ []string) error {
	projectId := flags.GetProjectOrFail()
	apiKey := flags.GetAPIKey()

	client := request.NewAPIClient(projectId, &apiKey)

	dests, err := getCanonicalDestinations(cmd.Context(), client)
	if err != nil {
		return fmt.Errorf("failed to get canonical destinations: %w", err)
	}

	if len(dests) == 0 {
		return fmt.Errorf("no valid destinations provided")
	}

	logger.Info("waiting for ngrok to start...")

	if err := waitForNgrok(cmd.Context()); err != nil {
		return fmt.Errorf("ngrok is not running: %w", err)
	}

	logger.Info("ngrok is running, fetching public URL...")

	publicURL, err := getPublicNgrokUrl(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get public ngrok URL: %w", err)
	}

	logger.Infof("Public ngrok URL: %s", publicURL)

	skipped := 0
	unchanged := 0
	updated := 0

	for _, dest := range dests {
		if dest.URL == publicURL {
			unchanged++
			continue
		}

		shouldUpdate, err := promptUpdateDestination(dest.String(), publicURL)
		if err != nil {
			return fmt.Errorf("failed to prompt for destination update: %w", err)
		}

		if !shouldUpdate {
			skipped++
			continue
		}

		logger.Infof("Changing webhook destination %s to %s", dest.Name, publicURL)

		_, err = client.PatchDestination(cmd.Context(), dest.Id, &request.PatchDestination{
			Destination: map[string]any{
				"metadata": map[string]any{
					"url": publicURL,
				},
			},
			UpdateMask: []string{"metadata.url"},
		})
		if err != nil {
			return fmt.Errorf("failed to update destination %s: %w", dest.String(), err)
		}

		updated++
	}

	logger.Infof("Total destinations: %d, unchanged: %d, skipped: %d, updated: %d",
		len(dests), unchanged, skipped, updated)

	return nil
}

func getCanonicalDestinations(ctx context.Context, client *request.APIClient) ([]Destination, error) {
	inputs := make([]string, 0, len(destinations))
	for _, dest := range destinations {
		if dest == "" {
			continue
		}

		inputs = append(inputs, dest)
	}

	if len(inputs) == 0 {
		return nil, fmt.Errorf("no valid destinations provided")
	}

	ampDests, err := client.ListDestinations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list destinations: %w", err)
	}

	ampNameMap := make(map[string]Destination)
	ampIdMap := make(map[string]Destination)

	for _, d := range ampDests {
		if d.Type != "webhook" {
			continue // Only consider webhook destinations
		}

		ampIdMap[strings.ToLower(d.Id)] = Destination{
			Id:   d.Id,
			Name: d.Name,
			URL:  d.Metadata.URL,
		}

		if d.Name != "" {
			ampNameMap[d.Name] = Destination{
				Id:   d.Id,
				Name: d.Name,
				URL:  d.Metadata.URL,
			}
		}
	}

	canonical := make([]Destination, 0, len(inputs))

	for _, input := range inputs {
		// Is it a known ID? If so add it.
		if dest, ok := ampIdMap[strings.ToLower(input)]; ok {
			canonical = append(canonical, dest)
			continue
		}

		// Is it a known name? If so, add it.
		if dest, ok := ampNameMap[input]; ok {
			canonical = append(canonical, dest)
			continue
		}

		return nil, fmt.Errorf("invalid destination: %s", input)
	}

	return canonical, nil
}

func promptUpdateDestination(dest, url string) (bool, error) {
	if skipConfirm {
		return true, nil
	}

	prompter := promptui.Prompt{
		Label:     "Change destination '" + dest + "' to '" + url + "'",
		IsConfirm: true,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	}

	_, err := prompter.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
