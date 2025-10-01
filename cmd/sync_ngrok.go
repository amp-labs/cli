package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	maxRetryDuration = 10 * time.Second
	initialDelay     = 500 * time.Millisecond
	maxDelay         = 2 * time.Second
	maxWaitDuration  = 5 * time.Minute
	retryInterval    = 2 * time.Second
	connectTimeout   = time.Second
)

var (
	// Static errors for linter compliance.
	errProtocolEmpty         = errors.New("protocol cannot be empty")
	errInvalidProtocol       = errors.New("invalid protocol: must be 'http' or 'https'")
	errNgrokServerEmpty      = errors.New("ngrok server cannot be empty")
	errNoTunnelsFound        = errors.New("no ngrok tunnels found")
	errNgrokAddressEmpty     = errors.New("ngrok server address cannot be empty")
	errNoDestinations        = errors.New("no valid destinations provided")
	errNotWebhook            = errors.New("destination is not a webhook")
	errInvalidDestination    = errors.New("invalid destination")
	errNgrokUnexpectedStatus = errors.New("ngrok API returned unexpected status code")
	errDeadlineExceeded      = errors.New("timeout waiting for ngrok")

	// destinations stores the list of destination names or IDs provided via command-line flags.
	// These can be either destination names (human-readable) or UUIDs (unique identifiers).
	destinations []string //nolint:gochecknoglobals

	// skipConfirm determines whether to skip user confirmation prompts when updating destinations.
	// When true, the command will automatically proceed with all updates without asking for confirmation.
	skipConfirm bool //nolint:gochecknoglobals

	// ngrokServer specifies the ngrok server URL for API calls.
	// Defaults to localhost:4040 but can be configured via command-line flag.
	ngrokServer string //nolint:gochecknoglobals

	// ngrokProtocol specifies the protocol to use when connecting to the ngrok server.
	// Defaults to "http" but can be configured via command-line flag to "http" or "https".
	ngrokProtocol string //nolint:gochecknoglobals

	// ngrokCommand defines the Cobra command structure for the ngrok subcommand.
	// This command is hidden from the main help output as it's primarily for development use.
	syncNgrokCommand = &cobra.Command{ //nolint:gochecknoglobals
		Use:    "sync-ngrok",
		Short:  "Sync Ampersand webhook destinations with ngrok tunnels",
		Long:   "Configure Ampersand destinations to use ngrok tunnels for local development.",
		Hidden: true, // Hidden because it's a development-only feature
		RunE:   runSyncNgrok,
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

// NameOrId returns the name of the destination if available,
// otherwise it falls back to the ID. This is useful for displaying
// destinations in user prompts or logs where a unique identifier is needed.
func (d Destination) NameOrId() string {
	// Return the name if available, otherwise fallback to ID
	if d.Name != "" {
		return d.Name
	}

	return d.Id
}

// ngrokTunnel represents a single tunnel from ngrok's API response.
// Contains the public URL that ngrok has assigned to forward traffic to the local service.
type ngrokTunnel struct {
	// PublicURL is the externally accessible URL that ngrok provides
	// (e.g., "https://abc123.ngrok.io")
	PublicURL string `json:"public_url"` //nolint:tagliatelle
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
	syncNgrokCommand.Flags().StringSliceVarP(&destinations,
		"destination", "D", []string{},
		"Destination names or IDs (UUIDs)")

	// Set up the confirmation skip flag for automated workflows
	syncNgrokCommand.Flags().BoolVarP(&skipConfirm,
		"yes", "y", false,
		"Skip confirmation prompts")

	// Set up the ngrok server flag for configurable ngrok API endpoint
	syncNgrokCommand.Flags().StringVarP(&ngrokServer,
		"ngrok-server", "s", "localhost:4040",
		"Ngrok server URL")

	// Set up the ngrok protocol flag for configurable protocol (http or https)
	syncNgrokCommand.Flags().StringVarP(&ngrokProtocol,
		"protocol", "P", "http",
		"Protocol to use for ngrok server (http or https)")

	// Register this command as a subcommand of the root CLI command
	rootCmd.AddCommand(syncNgrokCommand)
}

// validateProtocol validates that the ngrok protocol flag is set to a valid value.
// Returns an error if the protocol is not "http" or "https".
func validateProtocol() error {
	if ngrokProtocol == "" {
		return errProtocolEmpty
	}

	if ngrokProtocol != "http" && ngrokProtocol != "https" {
		return fmt.Errorf("%w: '%s'", errInvalidProtocol, ngrokProtocol)
	}

	return nil
}

// getPublicNgrokURLWithRetry retrieves the public URL of an active ngrok tunnel with retry logic.
// It handles the case where ngrok is running but still initializing (up to 10 seconds).
// Uses exponential backoff to reduce API load during initialization.
// Returns the selected tunnel's public URL or an error if retries are exhausted.
func getPublicNgrokURLWithRetry(ctx context.Context) (string, error) {
	if ngrokServer == "" {
		return "", errNgrokServerEmpty
	}

	deadline := time.Now().Add(maxRetryDuration)
	delay := initialDelay

	for {
		// Attempt to get the ngrok URL
		publicURL, err := getPublicNgrokURL(ctx)
		if err == nil {
			return publicURL, nil
		}
		// Check if we've exceeded the retry deadline
		if time.Now().After(deadline) {
			return "", fmt.Errorf("failed to get ngrok URL after %v: %w",
				maxRetryDuration.String(), err)
		}

		// Check if context was canceled
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(delay):
		}

		// Exponential backoff with cap
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}

// getPublicNgrokURL retrieves the public URL of an active ngrok tunnel by querying
// the ngrok local API endpoint. If multiple tunnels exist, it prompts the user to choose one.
// Returns the selected tunnel's public URL or an error if ngrok is not accessible.
func getPublicNgrokURL(ctx context.Context) (string, error) {
	// Create HTTP request to ngrok's API endpoint
	// The ngrok agent exposes a REST API, typically on port 4040 by default
	apiURL := url.URL{
		Scheme: ngrokProtocol,
		Host:   ngrokServer,
		Path:   "/api/tunnels",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
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
		return "", fmt.Errorf("%w %d", errNgrokUnexpectedStatus, resp.StatusCode)
	}

	// Parse the JSON response containing tunnel information
	var ngrokResp ngrokResponse

	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&ngrokResp); err != nil {
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
		return "", errNoTunnelsFound
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
		// User canceled selection or prompt failed
		return "", fmt.Errorf("failed to select tunnel: %w", err)
	}

	return urls[idx], nil
}

// waitForNgrok waits for the ngrok service to become available on the configured server.
// It performs an initial connectivity check, and if ngrok is not immediately available,
// it retries with a timeout and visual progress indicator (dots).
// Returns nil when ngrok is accessible, or an error if the timeout is exceeded.
func waitForNgrok(ctx context.Context) error {
	// Use configured ngrok server address
	address := ngrokServer
	if address == "" {
		return errNgrokAddressEmpty
	}

	// Perform initial connectivity check to see if ngrok is already running
	conn, err := net.DialTimeout("tcp", address, connectTimeout)
	if err == nil {
		// ngrok is already available, close connection and return immediately
		_ = conn.Close()

		return nil
	}

	// Set up retry loop with timeout and progress indication
	deadline := time.Now().Add(maxWaitDuration)
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
			// Context was canceled (e.g., user pressed Ctrl+C)
			// Return the context error (e.g., context.Canceled, context.DeadlineExceeded)
			return ctx.Err()
		case <-ticker.C:
			// Check if we've exceeded the maximum wait time
			if time.Now().After(deadline) {
				// Return timeout error with specific details for debugging
				return fmt.Errorf("%w on %s after %v",
					errDeadlineExceeded, address, maxWaitDuration)
			}

			// Attempt to connect to ngrok API endpoint
			conn, err = net.DialTimeout("tcp", address, connectTimeout)
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

// runSyncNgrok is the main entry point for the ngrok command execution.
// It orchestrates the entire process: setup, ngrok tunnel discovery, and destination updates.
// This function is called by the Cobra framework when the ngrok command is invoked.
func runSyncNgrok(cmd *cobra.Command, _ []string) error {
	// Validate protocol flag before proceeding
	if err := validateProtocol(); err != nil {
		return err
	}

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
		return nil, nil, errNoDestinations
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

	// Step 2: Query ngrok API for active tunnels with retry logic
	logger.Info("ngrok is running, fetching public URL...")

	publicURL, err := getPublicNgrokURLWithRetry(ctx)
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
func updateDestinations(ctx context.Context, client *request.APIClient,
	dests []Destination, publicURL string,
) destinationStats {
	// Initialize statistics tracking
	stats := destinationStats{}
	// Process each destination individually
	for _, dest := range dests {
		// Create the merged URL that would result from updating this destination
		mergedURL, err := mergeURLs(publicURL, dest.URL)
		if err != nil {
			// Log URL merge failures but continue with other destinations
			logger.Infof("failed to merge URLs for destination %s: %v",
				dest.NameOrId(), err)

			stats.skipped++

			continue
		}

		// Skip destinations that already have the correct URL
		if dest.URL == mergedURL {
			stats.unchanged++

			continue
		}

		// Ask user for confirmation (unless --yes flag is used)
		shouldUpdate, err := promptUpdateDestination(dest.Name, mergedURL)
		if err != nil {
			// Log prompt failures but continue with other destinations
			logger.Infof("failed to prompt for destination %s update: %v",
				dest.NameOrId(), err)

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
			logger.Infof("failed to update destination %s: %v", dest.NameOrId(), err)

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
// preserving all other destination configuration. The URL is constructed by merging
// the ngrok URL's protocol, host, and port with the existing URL's path and query parameters.
// Returns an error if the API call fails.
func updateDestination(ctx context.Context, client *request.APIClient, dest Destination, publicURL string) error {
	// Merge the ngrok URL with the existing destination URL to preserve path and query params
	mergedURL, err := mergeURLs(publicURL, dest.URL)
	if err != nil {
		return fmt.Errorf("failed to merge URLs: %w", err)
	}

	// Log the update operation for user visibility
	logger.Infof("Changing webhook destination %s to %s", dest.NameOrId(), mergedURL)

	// Make API call to update the destination's URL
	// Using PATCH with update mask to only modify the URL field
	_, err = client.PatchDestination(ctx, dest.Id, &request.PatchDestination{
		Destination: map[string]any{
			"metadata": map[string]any{
				"url": mergedURL,
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

// mergeURLs combines a new base URL (from ngrok) with the path and query parameters
// from an existing destination URL. This preserves the existing URL's path and query
// components while updating the protocol, host, and port to match the ngrok tunnel.
// Returns the merged URL string or an error if URL parsing fails.
func mergeURLs(ngrokURL, existingURL string) (string, error) {
	// Parse the ngrok URL to get the new base (protocol, host, port)
	baseURL, err := url.Parse(ngrokURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse ngrok URL '%s': %w", ngrokURL, err)
	}

	// Parse the existing destination URL to get path and query components
	existingParsed, err := url.Parse(existingURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse existing URL '%s': %w", existingURL, err)
	}

	// Create merged URL: use ngrok's scheme, host, and port with existing path and query
	mergedURL := &url.URL{
		Scheme:   baseURL.Scheme,
		Host:     baseURL.Host,
		Path:     existingParsed.Path,
		RawQuery: existingParsed.RawQuery,
		Fragment: existingParsed.Fragment,
	}

	return mergedURL.String(), nil
}

// getCanonicalDestinations resolves user-provided destination identifiers (names or IDs)
// into canonical Destination objects by querying the Ampersand API. It supports matching
// by both destination name and UUID, and only includes webhook-type destinations.
// Returns a slice of resolved destinations or an error if any identifier is invalid.
func getCanonicalDestinations(ctx context.Context, client *request.APIClient) ([]Destination, error) {
	inputs := filterEmptyDestinations()
	if len(inputs) == 0 {
		return nil, errNoDestinations
	}

	ampDests, err := client.ListDestinations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list destinations: %w", err)
	}

	ampNameMap, ampIdMap := buildDestinationMaps(ampDests)

	return resolveInputsToDestinations(inputs, ampNameMap, ampIdMap)
}

// filterEmptyDestinations removes empty strings from the destinations slice.
func filterEmptyDestinations() []string {
	inputs := make([]string, 0, len(destinations))

	for _, dest := range destinations {
		if dest != "" {
			inputs = append(inputs, dest)
		}
	}

	return inputs
}

// buildDestinationMaps creates lookup maps for efficient destination resolution.
func buildDestinationMaps(ampDests []*request.Destination) (map[string]Destination, map[string]Destination) {
	ampNameMap := make(map[string]Destination)
	ampIdMap := make(map[string]Destination)

	for _, destination := range ampDests {
		dest := Destination{
			Id:   destination.Id,
			Name: destination.Name,
			URL:  destination.Metadata.URL,
			Type: destination.Type,
		}

		ampIdMap[strings.ToLower(destination.Id)] = dest

		if destination.Name != "" {
			ampNameMap[strings.ToLower(destination.Name)] = dest
		}
	}

	return ampNameMap, ampIdMap
}

// resolveInputsToDestinations resolves user inputs to canonical destination objects.
func resolveInputsToDestinations(inputs []string, ampNameMap, ampIdMap map[string]Destination) ([]Destination, error) {
	canonical := make([]Destination, 0, len(inputs))

	for _, input := range inputs {
		dest, err := resolveDestination(input, ampNameMap, ampIdMap)
		if err != nil {
			return nil, err
		}

		canonical = append(canonical, dest)
	}

	return canonical, nil
}

// resolveDestination resolves a single input to a destination.
func resolveDestination(input string, ampNameMap, ampIdMap map[string]Destination) (Destination, error) {
	// Try to match by ID first (case-insensitive)
	if dest, ok := ampIdMap[strings.ToLower(input)]; ok {
		return validateWebhookDestination(dest, input)
	}

	// Try to match by name (exact match)
	if dest, ok := ampNameMap[strings.ToLower(input)]; ok {
		return validateWebhookDestination(dest, input)
	}

	return Destination{}, fmt.Errorf("%w: %s", errInvalidDestination, input)
}

// validateWebhookDestination validates that a destination is a webhook.
func validateWebhookDestination(dest Destination, input string) (Destination, error) {
	if dest.Type != "webhook" {
		return Destination{}, fmt.Errorf("%w: %s", errNotWebhook, input)
	}

	return dest, nil
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
	var labelBuilder strings.Builder

	labelBuilder.WriteString("Change destination '")
	labelBuilder.WriteString(dest)
	labelBuilder.WriteString("' to '")
	labelBuilder.WriteString(url)
	labelBuilder.WriteString("'")

	prompter := promptui.Prompt{
		Label:     labelBuilder.String(),
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
