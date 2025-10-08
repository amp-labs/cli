package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/amp-labs/cli/internal/webhook"
	"github.com/amp-labs/cli/logger"
	"github.com/spf13/cobra"
)

var (
	ErrFailedToGetTCPAddress = errors.New("failed to get TCP address")
	forwardURL               string
	listenAddr               string
	listenCommand            = &cobra.Command{
		Use:   "listen",
		Short: "Listen for webhooks locally",
		Long: `Listen for webhooks locally and forward them to your application.
This command starts a local webhook server that receives events and forwards them to your application.
It's designed for local development and testing.`,
		Hidden: true,
		RunE:   runListen,
	}
)

func init() {
	listenCommand.Flags().StringVar(&forwardURL, "forward-to", "http://localhost:4000/webhook",
		"URL to forward webhooks to")
	listenCommand.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:0", "Address to listen on (default is random port)")
	rootCmd.AddCommand(listenCommand)
}

func runListen(cmd *cobra.Command, args []string) error {
	// Set up a context that we can cancel
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Set up the HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleWebhook)

	// Start the server
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	// Save the port for the trigger command to use
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return ErrFailedToGetTCPAddress
	}

	port := strconv.Itoa(addr.Port)

	// Save the port
	if err := saveListenerPort(port); err != nil {
		logger.FatalErr("failed to save port", err)
	}

	const serverTimeout = 10 * time.Second
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: serverTimeout,
	}

	// Start the server in a goroutine so it doesn't block
	go func() {
		logger.Info("starting webhook listener")

		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.FatalErr("webhook listener failed", err)
		}
	}()

	// Print the listen address
	fmt.Fprint(os.Stdout, "🎧 Listening on "+addr.IP.String()+":"+port+"\n")
	fmt.Fprint(os.Stdout, "ℹ️  Forwarding to: "+forwardURL+"\n")
	fmt.Fprint(os.Stdout, "Press Ctrl+C to stop\n")

	// Wait for interrupt signal
	<-ctx.Done()
	fmt.Fprint(os.Stdout, "\nShutting down webhook listener...\n")

	// Create a deadline to wait for current connections to complete
	const shutdownTimeout = 5 * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.FatalErr("shutdown error", err)
	}

	// Clean up the port file
	clearListenerPort()

	fmt.Fprint(os.Stdout, "Webhook listener stopped\n")

	return nil
}

const (
	mkdirPerm = 0o700
	filePerm  = 0o600
)

// saveListenerPort saves the listener port to a temporary file.
func saveListenerPort(port string) error {
	dir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	// Create ampersand directory if it doesn't exist
	ampDir := filepath.Join(dir, "ampersand")

	if err := os.MkdirAll(ampDir, mkdirPerm); err != nil {
		return err
	}

	// Write port to file
	portFile := filepath.Join(ampDir, "webhook-port")

	return os.WriteFile(portFile, []byte(port), filePerm)
}

// clearListenerPort removes the port file.
func clearListenerPort() {
	dir, err := os.UserCacheDir()
	if err != nil {
		return
	}

	portFile := filepath.Join(dir, "ampersand", "webhook-port")
	os.Remove(portFile) // Ignore errors
}

func handleWebhook(writer http.ResponseWriter, req *http.Request) {
	// Only accept POST requests
	if req.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Read the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.FatalErr("error reading request body", err)
		http.Error(writer, "Bad request", http.StatusBadRequest)

		return
	}

	req.Body.Close()

	// Log the webhook payload
	if err := webhook.PrettyPrintJSON(body); err != nil {
		logger.FatalErr("error pretty printing JSON", err)
	}

	// Forward the request to the application
	forwardReq, err := http.NewRequestWithContext(req.Context(), http.MethodPost, forwardURL, bytes.NewReader(body))
	if err != nil {
		logger.FatalErr("error creating forward request", err)
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Copy headers
	for k, v := range req.Header {
		if k == "Content-Length" {
			continue
		}

		for _, vv := range v {
			forwardReq.Header.Add(k, vv)
		}
	}

	// Forward the request
	const clientTimeout = 5 * time.Second
	client := &http.Client{Timeout: clientTimeout}

	resp, err := client.Do(forwardReq)
	if err != nil {
		logger.FatalErr("error forwarding request to "+forwardURL, err)
		// Still return 200 to the original sender
		writer.WriteHeader(http.StatusOK)

		return
	}

	defer resp.Body.Close()

	// Copy the response from the application back to the original sender
	for k, v := range resp.Header {
		for _, vv := range v {
			writer.Header().Add(k, vv)
		}
	}

	writer.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		logger.FatalErr("error copying response", err)
	}
}
