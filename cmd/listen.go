package cmd

import (
	"bytes"
	"context"
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
	forwardURL    string
	listenAddr    string
	listenCommand = &cobra.Command{
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
	listenCommand.Flags().StringVar(&forwardURL, "forward-to", "http://localhost:4000/webhook", "URL to forward webhooks to")
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
	addr := listener.Addr().(*net.TCPAddr)
	port := strconv.Itoa(addr.Port)

	// Save the port
	if err := saveListenerPort(port); err != nil {
		logger.FatalErr("failed to save port", err)
	}

	srv := &http.Server{
		Handler: mux,
	}

	// Start the server in a goroutine so it doesn't block
	go func() {
		logger.Info("starting webhook listener")

		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.FatalErr("webhook listener failed", err)
		}
	}()

	// Print the listen address
	fmt.Printf("🎧 Listening on %s:%s\n", addr.IP, port)
	fmt.Printf("ℹ️  Forwarding to: %s\n", forwardURL)
	fmt.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	<-ctx.Done()
	fmt.Println("\nShutting down webhook listener...")

	// Create a deadline to wait for current connections to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.FatalErr("shutdown error", err)
	}

	// Clean up the port file
	clearListenerPort()

	fmt.Println("Webhook listener stopped")

	return nil
}

// saveListenerPort saves the listener port to a temporary file
func saveListenerPort(port string) error {
	dir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	// Create ampersand directory if it doesn't exist
	ampDir := filepath.Join(dir, "ampersand")
	if err := os.MkdirAll(ampDir, 0700); err != nil {
		return err
	}

	// Write port to file
	portFile := filepath.Join(ampDir, "webhook-port")

	return os.WriteFile(portFile, []byte(port), 0600)
}

// clearListenerPort removes the port file
func clearListenerPort() {
	dir, err := os.UserCacheDir()
	if err != nil {
		return
	}

	portFile := filepath.Join(dir, "ampersand", "webhook-port")
	os.Remove(portFile) // Ignore errors
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.FatalErr("error reading request body", err)
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	r.Body.Close()

	// Log the webhook payload
	webhook.PrettyPrintJSON(body)

	// Forward the request to the application
	req, err := http.NewRequest(http.MethodPost, forwardURL, bytes.NewReader(body))
	if err != nil {
		logger.FatalErr("error creating forward request", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)

		return
	}

	// Copy headers
	for k, v := range r.Header {
		if k == "Content-Length" {
			continue
		}

		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	// Forward the request
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		logger.FatalErr(fmt.Sprintf("error forwarding request to %s", forwardURL), err)
		// Still return 200 to the original sender
		w.WriteHeader(http.StatusOK)

		return
	}

	defer resp.Body.Close()

	// Copy the response from the application back to the original sender
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		logger.FatalErr("error copying response", err)
	}
}
