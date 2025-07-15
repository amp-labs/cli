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
	// destinations stores the list of destination names or IDs provided via command-line flags.
	// These can be either destination names (human-readable) or UUIDs (unique identifiers).
	destinations []string

	// skipConfirm determines whether to skip user confirmation prompts when updating destinations.
	// When true, the command will automatically proceed with all updates without asking for confirmation.
	skipConfirm bool

	// ngrokCommand defines the Cobra command structure for the ngrok subcommand.
	// This command is hidden from the main help output as it's primarily for development use.
	ngrokCommand = &cobra.Command{
		Use:    "ngrok",
		Short:  "Automatic ngrok URL configuration for Ampersand destinations",
		Long:   "Configure Ampersand destinations to use ngrok tunnels for local development.",
		Hidden: true, // Hidden because it's a development-only feature
		RunE:   runNgrok,
	}
)

// Destination represents a webhook destination in the Ampersand platform.
// It contains the essential information needed to identify and update a destination's URL.
type Destination struct {
	// Id is the unique UUID identifier for the destination in Ampersand
	Id string `json:"id"`

	// Name is the human-readable name assigned to the destination (optional)
	Name string `json:"name"`

	// URL is the current webhook URL where events are sent
	URL string `json:"url"`

	// Type indicates the kind of destination, typically "webhook" for webhook destinations
	Type string `json:"type"` // Type of the destination (e.g., "webhook")
}

// String returns a human-readable representation of the destination.
// If a name is available, it returns "Name (ID)", otherwise just the ID.
// This method implements the Stringer interface for better display in prompts and logs.
func (d Destination) String() string {
	// Prefer showing name with ID for clarity, fallback to just ID
	if d.Name != "" {
		return fmt.Sprintf("%s (%s)", d.Name, d.Id)
	}

	return d.Id
}

// ngrokTunnel represents a single tunnel from ngrok's API response.
// Contains the public URL that ngrok has assigned to forward traffic to the local service.
type ngrokTunnel struct {
	// PublicURL is the externally accessible URL that ngrok provides
	// (e.g., "https://abc123.ngrok.io")
	PublicURL string `json:"public_url"`
}

// ngrokResponse represents the JSON response from ngrok's local API endpoint.
// The ngrok agent exposes an API at localhost:4040 that provides information about active tunnels.
type ngrokResponse struct {
	// Tunnels is an array of all currently active ngrok tunnels
	Tunnels []ngrokTunnel `json:"tunnels"`
}

// init initializes the ngrok command by setting up command-line flags and
// registering the command with the root command. This function is called
// automatically when the package is imported.
func init() {
	// Set up the destination flag to accept multiple destination identifiers
	// Users can specify destinations by name or UUID, multiple times or comma-separated
	ngrokCommand.Flags().StringSliceVarP(&destinations,
		"destination", "D", []string{},
		"Destination names or IDs (UUIDs)")

	// Set up the confirmation skip flag for automated workflows
	ngrokCommand.Flags().BoolVarP(&skipConfirm,
		"yes", "y", false,
		"Skip confirmation prompts")

	// Register this command as a subcommand of the root CLI command
	rootCmd.AddCommand(ngrokCommand)
}

// getPublicNgrokUrl retrieves the public URL of an active ngrok tunnel by querying
// the ngrok local API endpoint. If multiple tunnels exist, it prompts the user to choose one.
// Returns the selected tunnel's public URL or an error if ngrok is not accessible.
func getPublicNgrokUrl(ctx context.Context) (string, error) {
	// Create HTTP request to ngrok's local API endpoint
	// The ngrok agent exposes a REST API on port 4040 by default
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:4040/api/tunnels", nil)
	if err != nil {
		// Wrap the error with context about what operation failed
		return "", fmt.Errorf("failed to create ngrok API request: %w", err)
	}

	// Execute the HTTP request to get tunnel information
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// This error typically means ngrok is not running or not accessible
		return "", fmt.Errorf("failed to connect to ngrok API: %w", err)
	}

	// Ensure response body is properly closed to prevent resource leaks
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error closing response body: %v\n", closeErr)
		}
	}()

	// Check if the ngrok API responded successfully
	if resp.StatusCode != http.StatusOK {
		// Non-200 status codes indicate ngrok API issues
		return "", fmt.Errorf("ngrok API returned status %d", resp.StatusCode)
	}

	// Parse the JSON response containing tunnel information
	var ngrokResp ngrokResponse
	if err := json.NewDecoder(resp.Body).Decode(&ngrokResp); err != nil {
		// JSON parsing errors indicate malformed response from ngrok
		return "", fmt.Errorf("failed to parse ngrok response: %w", err)
	}

	// Handle tunnel selection (single tunnel vs. multiple tunnels)
	return chooseNgrokTunnel(&ngrokResp)
}

// chooseNgrokTunnel handles tunnel selection when multiple ngrok tunnels are active.
// If only one tunnel exists, it returns that tunnel's URL automatically.
// If multiple tunnels exist, it presents an interactive prompt for the user to choose.
// Returns the selected tunnel's public URL or an error if selection fails.
func chooseNgrokTunnel(ngrokResp *ngrokResponse) (string, error) {
	// Validate that at least one tunnel is available
	if len(ngrokResp.Tunnels) == 0 {
		// This means ngrok is running but no tunnels are active
		return "", fmt.Errorf("no ngrok tunnels found")
	}

	// If only one tunnel exists, use it automatically (no need to prompt)
	if len(ngrokResp.Tunnels) == 1 {
		return ngrokResp.Tunnels[0].PublicURL, nil
	}

	// Extract URLs from tunnels for user selection
	urls := make([]string, len(ngrokResp.Tunnels))
	for i, tunnel := range ngrokResp.Tunnels {
		urls[i] = tunnel.PublicURL
	}

	// Present interactive selection prompt for multiple tunnels
	prompt := promptui.Select{
		Label:  "Choose ngrok tunnel",
		Items:  urls,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	// Get user's tunnel selection
	idx, _, err := prompt.Run()
	if err != nil {
		// User cancelled selection or prompt failed
		return "", fmt.Errorf("failed to select tunnel: %w", err)
	}

	return urls[idx], nil
}

// waitForNgrok waits for the ngrok service to become available on localhost:4040.
// It performs an initial connectivity check, and if ngrok is not immediately available,
// it retries with a timeout and visual progress indicator (dots).
// Returns nil when ngrok is accessible, or an error if the timeout is exceeded.
func waitForNgrok(ctx context.Context) error {
	// Configuration constants for ngrok availability checking
	const (
		maxDuration   = 5 * time.Minute  // Maximum time to wait for ngrok
		retryInterval = 2 * time.Second  // Time between connection attempts
		address       = "localhost:4040" // Standard ngrok API address
	)

	// Perform initial connectivity check to see if ngrok is already running
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err == nil {
		// ngrok is already available, close connection and return immediately
		_ = conn.Close()
		return nil
	}

	// Set up retry loop with timeout and progress indication
	deadline := time.Now().Add(maxDuration)
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	// Track if we need to add a newline after progress dots
	needsNewline := false

	// Ensure we clean up the progress display when function exits
	defer func() {
		if needsNewline {
			_, _ = os.Stdout.Write([]byte("\n"))
		}
	}()

	// Main retry loop with context cancellation and timeout handling
	for {
		select {
		case <-ctx.Done():
			// Context was cancelled (e.g., user pressed Ctrl+C)
			// Return the context error (e.g., context.Canceled, context.DeadlineExceeded)
			return ctx.Err()
		case <-ticker.C:
			// Check if we've exceeded the maximum wait time
			if time.Now().After(deadline) {
				// Return timeout error with specific details for debugging
				return fmt.Errorf("timeout waiting for ngrok on %s after %v", address, maxDuration)
			}

			// Attempt to connect to ngrok API endpoint
			conn, err := net.DialTimeout("tcp", address, time.Second)
			if err == nil {
				// Success! ngrok is now available
				_ = conn.Close()
				return nil
			}

			// Show progress indicator (dot) and continue waiting
			_, _ = os.Stdout.Write([]byte("."))
			_ = os.Stdout.Sync()
			needsNewline = true
		}
	}
}

// runNgrok is the main entry point for the ngrok command execution.
// It orchestrates the entire process: setup, ngrok tunnel discovery, and destination updates.
// This function is called by the Cobra framework when the ngrok command is invoked.
func runNgrok(cmd *cobra.Command, _ []string) error {
	// Phase 1: Initialize API client and resolve destination identifiers
	client, dests, err := setupNgrokExecution(cmd.Context())
	if err != nil {
		// Setup errors are passed through directly as they already have context
		return err
	}

	// Phase 2: Wait for ngrok and get the public tunnel URL
	publicURL, err := getNgrokTunnelURL(cmd.Context())
	if err != nil {
		// Ngrok-related errors are passed through with their original context
		return err
	}

	// Phase 3: Update all specified destinations with the ngrok URL
	stats := updateDestinations(cmd.Context(), client, dests, publicURL)

	// Phase 4: Report the results to the user
	logDestinationStats(stats, len(dests))

	return nil
}

// setupNgrokExecution initializes the necessary components for ngrok command execution.
// It creates an API client with the current project context and resolves the provided
// destination identifiers to canonical destination objects.
// Returns the API client, resolved destinations, and any setup errors.
func setupNgrokExecution(ctx context.Context) (*request.APIClient, []Destination, error) {
	// Get the current project context and API credentials
	projectId := flags.GetProjectOrFail()
	apiKey := flags.GetAPIKey()
	client := request.NewAPIClient(projectId, &apiKey)

	// Resolve user-provided destination identifiers to actual destination objects
	dests, err := getCanonicalDestinations(ctx, client)
	if err != nil {
		// Add context about this specific operation failure
		return nil, nil, fmt.Errorf("failed to get canonical destinations: %w", err)
	}

	// Validate that we have at least one destination to work with
	if len(dests) == 0 {
		// This indicates user provided invalid destination identifiers
		return nil, nil, fmt.Errorf("no valid destinations provided")
	}

	return client, dests, nil
}

// getNgrokTunnelURL manages the process of waiting for ngrok to start and retrieving
// the public tunnel URL. It provides user feedback during the wait process and
// handles the tunnel URL selection if multiple tunnels are available.
// Returns the selected ngrok public URL or an error if the process fails.
func getNgrokTunnelURL(ctx context.Context) (string, error) {
	// Step 1: Wait for ngrok service to become available
	logger.Info("waiting for ngrok to start...")

	if err := waitForNgrok(ctx); err != nil {
		// Provide helpful context about ngrok not being available
		return "", fmt.Errorf("ngrok is not running: %w", err)
	}

	// Step 2: Query ngrok API for active tunnels and get public URL
	logger.Info("ngrok is running, fetching public URL...")

	publicURL, err := getPublicNgrokUrl(ctx)
	if err != nil {
		// Add context about URL retrieval failure
		return "", fmt.Errorf("failed to get public ngrok URL: %w", err)
	}

	// Step 3: Display the selected ngrok URL
	logger.Infof("Public ngrok URL: %s", publicURL)

	return publicURL, nil
}

// destinationStats tracks the outcome of destination update operations.
// It provides a summary of how many destinations were processed in each category.
type destinationStats struct {
	// skipped counts destinations that were not updated (user declined or errors occurred)
	skipped int

	// unchanged counts destinations that already had the correct ngrok URL
	unchanged int

	// updated counts destinations that were successfully updated with the new ngrok URL
	updated int
}

// updateDestinations processes a list of destinations and updates their URLs to point
// to the provided ngrok public URL. For each destination, it checks if an update is needed,
// prompts for user confirmation (unless skipped), and performs the update via the API.
// Returns statistics about the update operation (updated, skipped, unchanged counts).
func updateDestinations(ctx context.Context, client *request.APIClient, dests []Destination, publicURL string) destinationStats {
	// Initialize statistics tracking
	stats := destinationStats{}

	// Process each destination individually
	for _, dest := range dests {
		// Skip destinations that already have the correct URL
		if dest.URL == publicURL {
			stats.unchanged++
			continue
		}

		// Ask user for confirmation (unless --yes flag is used)
		shouldUpdate, err := promptUpdateDestination(dest.String(), publicURL)
		if err != nil {
			// Log prompt failures but continue with other destinations
			logger.Info(fmt.Sprintf("failed to prompt for destination update: %v", err))
			stats.skipped++
			continue
		}

		// Respect user's decision to skip this destination
		if !shouldUpdate {
			stats.skipped++
			continue
		}

		// Attempt to update the destination via API
		if err := updateDestination(ctx, client, dest, publicURL); err != nil {
			// Log API update failures but continue with other destinations
			logger.Info(fmt.Sprintf("failed to update destination %s: %v", dest.String(), err))
			stats.skipped++
			continue
		}

		// Successfully updated this destination
		stats.updated++
	}

	return stats
}

// updateDestination performs the actual API call to update a single destination's URL.
// It uses the PATCH endpoint to modify only the metadata.url field of the destination,
// preserving all other destination configuration.
// Returns an error if the API call fails.
func updateDestination(ctx context.Context, client *request.APIClient, dest Destination, publicURL string) error {
	// Log the update operation for user visibility
	logger.Infof("Changing webhook destination %s to %s", dest.Name, publicURL)

	// Make API call to update the destination's URL
	// Using PATCH with update mask to only modify the URL field
	_, err := client.PatchDestination(ctx, dest.Id, &request.PatchDestination{
		Destination: map[string]any{
			"metadata": map[string]any{
				"url": publicURL,
			},
		},
		UpdateMask: []string{"metadata.url"}, // Only update the URL field
	})

	return err
}

// logDestinationStats outputs a summary of the destination update operation.
// It provides a clear breakdown of how many destinations were processed in each category,
// helping users understand the results of the ngrok command execution.
func logDestinationStats(stats destinationStats, total int) {
	logger.Infof("Total destinations: %d, unchanged: %d, skipped: %d, updated: %d",
		total, stats.unchanged, stats.skipped, stats.updated)
}

// getCanonicalDestinations resolves user-provided destination identifiers (names or IDs)
// into canonical Destination objects by querying the Ampersand API. It supports matching
// by both destination name and UUID, and only includes webhook-type destinations.
// Returns a slice of resolved destinations or an error if any identifier is invalid.
func getCanonicalDestinations(ctx context.Context, client *request.APIClient) ([]Destination, error) {
	// Filter out empty destination strings from command-line input
	inputs := make([]string, 0, len(destinations))
	for _, dest := range destinations {
		if dest == "" {
			continue // Skip empty strings
		}

		inputs = append(inputs, dest)
	}

	// Validate that we have at least one destination to process
	if len(inputs) == 0 {
		// This means user didn't provide any destination flags or all were empty
		return nil, fmt.Errorf("no valid destinations provided")
	}

	// Fetch all destinations from the Ampersand API
	ampDests, err := client.ListDestinations(ctx)
	if err != nil {
		// API call failed - could be network, auth, or server issues
		return nil, fmt.Errorf("failed to list destinations: %w", err)
	}

	// Create lookup maps for efficient destination resolution
	// Support both name-based and ID-based lookups
	ampNameMap := make(map[string]Destination)
	ampIdMap := make(map[string]Destination)

	// Build lookup maps from API response
	for _, d := range ampDests {
		// Add to ID-based lookup (case-insensitive for UUIDs)
		ampIdMap[strings.ToLower(d.Id)] = Destination{
			Id:   d.Id,
			Name: d.Name,
			URL:  d.Metadata.URL,
			Type: d.Type,
		}

		// Add to name-based lookup if destination has a name
		if d.Name != "" {
			ampNameMap[d.Name] = Destination{
				Id:   d.Id,
				Name: d.Name,
				URL:  d.Metadata.URL,
				Type: d.Type,
			}
		}
	}

	// Resolve each input to a canonical destination object
	canonical := make([]Destination, 0, len(inputs))

	for _, input := range inputs {
		// Try to match by ID first (case-insensitive)
		if dest, ok := ampIdMap[strings.ToLower(input)]; ok {
			if dest.Type != "webhook" {
				// If the destination is not a webhook, skip it
				return nil, fmt.Errorf("destination %s is not a webhook", input)
			}

			canonical = append(canonical, dest)
			continue
		}

		// Try to match by name (exact match)
		if dest, ok := ampNameMap[input]; ok {
			if dest.Type != "webhook" {
				// If the destination is not a webhook, skip it
				return nil, fmt.Errorf("destination %s is not a webhook", input)
			}

			canonical = append(canonical, dest)
			continue
		}

		// Input doesn't match any known destination name or ID
		// Provide specific error to help user identify the issue
		return nil, fmt.Errorf("invalid destination: %s", input)
	}

	return canonical, nil
}

// promptUpdateDestination asks the user for confirmation before updating a destination's URL.
// If the skipConfirm flag is set, it automatically returns true without prompting.
// Otherwise, it presents an interactive yes/no prompt to the user.
// Returns true if the user confirms the update, false if declined, or an error if the prompt fails.
func promptUpdateDestination(dest, url string) (bool, error) {
	// If skip confirmation flag is set, automatically approve all updates
	if skipConfirm {
		return true, nil
	}

	// Present interactive confirmation prompt to user
	prompter := promptui.Prompt{
		Label:     "Change destination '" + dest + "' to '" + url + "'",
		IsConfirm: true, // Makes this a yes/no prompt
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
	}

	// Execute the prompt and handle user response
	_, err := prompter.Run()
	if err != nil {
		// User pressed Ctrl+C or said "no" - not an error condition
		if errors.Is(err, promptui.ErrAbort) {
			return false, nil
		}

		// Actual error occurred during prompting (e.g., stdin issues)
		return false, err
	}

	// User confirmed the update
	return true, nil
}
