package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/amp-labs/cli/flags"
	"github.com/amp-labs/cli/logger"
	"github.com/amp-labs/cli/request"
	"github.com/spf13/cobra"
)

const (
	defaultTunnelTarget = "http://127.0.0.1:4000"
	tunnelStartTimeout  = 30 * time.Second
	tunnelStopTimeout   = 10 * time.Second
)

var (
	errCloudflaredNotFound = errors.New("cloudflared was not found in PATH")
	errTunnelURLTimeout    = errors.New("timed out waiting for the Cloudflare tunnel URL")
	errTunnelStopped       = errors.New("cloudflared stopped before the tunnel was ready")
	cloudflareURLPattern   = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)
	tunnelTarget           string
	tunnelCommand          = &cobra.Command{
		Use:    "tunnel <destination>",
		Short:  "Start a local tunnel for a webhook destination",
		Long:   "Start a Cloudflare tunnel, point one webhook destination to it, and restore the destination when stopped.",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE:   runTunnelCommand,
	}
)

type runningTunnel struct {
	publicURL string
	done      <-chan error
	stop      context.CancelFunc
}

type tunnelStarter func(context.Context, string) (*runningTunnel, error)

func init() {
	tunnelCommand.Flags().StringVar(&tunnelTarget, "target", defaultTunnelTarget,
		"Local URL that receives tunneled requests")
	rootCmd.AddCommand(tunnelCommand)
}

func runTunnelCommand(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	projectId := flags.GetProjectOrFail()
	apiKey := flags.GetAPIKey()
	client := request.NewAPIClient(projectId, &apiKey)

	destination, err := getTunnelDestination(ctx, client, args[0])
	if err != nil {
		return err
	}

	return runDestinationTunnel(ctx, client, destination, tunnelTarget, startCloudflareTunnel)
}

func getTunnelDestination(
	ctx context.Context, client *request.APIClient, identifier string,
) (Destination, error) {
	destinations, err := client.ListDestinations(ctx)
	if err != nil {
		return Destination{}, fmt.Errorf("list destinations: %w", err)
	}

	nameMap, idMap := buildDestinationMaps(destinations)

	return resolveDestination(identifier, nameMap, idMap)
}

func runDestinationTunnel(
	ctx context.Context,
	client *request.APIClient,
	destination Destination,
	targetURL string,
	start tunnelStarter,
) error {
	tunnel, err := start(ctx, targetURL)
	if err != nil {
		return err
	}

	if tunnel.stop != nil {
		defer tunnel.stop()
	}

	temporaryURL, err := mergeURLs(tunnel.publicURL, destination.URL)
	if err != nil {
		return err
	}

	err = ctx.Err()
	if err != nil {
		return err
	}

	updateCtx, cancelUpdate := context.WithTimeout(context.WithoutCancel(ctx), tunnelStopTimeout)
	err = setDestinationURL(updateCtx, client, destination.Id, temporaryURL)

	cancelUpdate()

	if err != nil {
		return err
	}

	logger.Infof("Tunnel ready: %s", temporaryURL)
	logger.Info("Press Ctrl+C to stop and restore the destination")

	var tunnelErr error

	select {
	case <-ctx.Done():
	case tunnelErr = <-tunnel.done:
	}

	restoreCtx, cancelRestore := context.WithTimeout(context.WithoutCancel(ctx), tunnelStopTimeout)
	defer cancelRestore()

	err = setDestinationURL(restoreCtx, client, destination.Id, destination.URL)
	if err != nil {
		return fmt.Errorf("restore destination URL: %w", err)
	}

	logger.Infof("Restored destination: %s", destination.URL)

	if ctx.Err() == nil {
		if tunnelErr == nil {
			tunnelErr = errTunnelStopped
		}

		return fmt.Errorf("cloudflared stopped: %w", tunnelErr)
	}

	return nil
}

func setDestinationURL(
	ctx context.Context, client *request.APIClient, destinationId string, destinationURL string,
) error {
	_, err := client.PatchDestination(ctx, destinationId, &request.PatchDestination{
		Destination: map[string]any{
			"metadata": map[string]any{
				"url": destinationURL,
			},
		},
		UpdateMask: []string{"metadata.url"},
	})
	if err != nil {
		return fmt.Errorf("update destination URL: %w", err)
	}

	return nil
}

func startCloudflareTunnel(ctx context.Context, targetURL string) (*runningTunnel, error) {
	cloudflaredPath, err := exec.LookPath("cloudflared")
	if err != nil {
		return nil, errCloudflaredNotFound
	}

	processCtx, stop := context.WithCancel(ctx)
	// cloudflaredPath is resolved from PATH, and targetURL is passed as one argument without a shell.
	//nolint:gosec
	command := exec.CommandContext(processCtx, cloudflaredPath,
		"tunnel", "--url", targetURL, "--no-autoupdate")
	command.Stdout = io.Discard

	stderr, err := command.StderrPipe()
	if err != nil {
		stop()

		return nil, fmt.Errorf("read cloudflared output: %w", err)
	}

	err = command.Start()
	if err != nil {
		stop()

		return nil, fmt.Errorf("start cloudflared: %w", err)
	}

	lines := scanLines(stderr)
	done := make(chan error, 1)

	go func() {
		done <- command.Wait()
	}()

	publicURL, err := waitForCloudflareTunnelURL(processCtx, lines, done)
	if err != nil {
		stop()

		return nil, err
	}

	go discardTunnelOutput(lines)

	return &runningTunnel{
		publicURL: publicURL,
		done:      done,
		stop:      stop,
	}, nil
}

func scanLines(reader io.Reader) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
	}()

	return lines
}

func waitForCloudflareTunnelURL(
	ctx context.Context, lines <-chan string, done <-chan error,
) (string, error) {
	timer := time.NewTimer(tunnelStartTimeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timer.C:
			return "", errTunnelURLTimeout
		case err := <-done:
			if err == nil {
				return "", errTunnelStopped
			}

			return "", fmt.Errorf("%w: %w", errTunnelStopped, err)
		case line, ok := <-lines:
			if !ok {
				lines = nil

				continue
			}

			if publicURL := cloudflareTunnelURL(line); publicURL != "" {
				return publicURL, nil
			}
		}
	}
}

func cloudflareTunnelURL(line string) string {
	return cloudflareURLPattern.FindString(line)
}

func discardTunnelOutput(lines <-chan string) {
	for range lines {
	}
}
