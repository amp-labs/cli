package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestLoginCommandUsesHeadlessRunner(t *testing.T) {
	t.Parallel()

	loggedOut := false
	openedBrowser := false
	ranHeadless := false
	cmd := newLoginCmd(
		func(showLogs bool) {
			if showLogs {
				t.Fatal("logout logs were enabled")
			}

			loggedOut = true
		},
		func() { openedBrowser = true },
		func(context.Context) { ranHeadless = true },
	)
	cmd.SetArgs([]string{"--headless"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute login command: %v", err)
	}

	if !loggedOut {
		t.Fatal("existing login was not cleared")
	}

	if openedBrowser {
		t.Fatal("browser login ran in headless mode")
	}

	if !ranHeadless {
		t.Fatal("headless login did not run")
	}
}

func TestLoginCommandUsesBrowserRunnerByDefault(t *testing.T) {
	t.Parallel()

	openedBrowser := false
	ranHeadless := false
	cmd := newLoginCmd(
		func(bool) {},
		func() { openedBrowser = true },
		func(context.Context) { ranHeadless = true },
	)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("execute login command: %v", err)
	}

	if !openedBrowser {
		t.Fatal("browser login did not run")
	}

	if ranHeadless {
		t.Fatal("headless login ran without --headless")
	}
}

func TestParseLoginCallback(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"cookies":{"session":"test"}}`)
	callback := fmt.Sprintf(
		"http://localhost:%d/done?p=%s",
		ServerPort,
		base64.StdEncoding.EncodeToString(payload),
	)

	got, err := parseLoginCallback("  " + callback + "\n")
	if err != nil {
		t.Fatalf("parse callback: %v", err)
	}

	if string(got) != string(payload) {
		t.Fatalf("payload = %q, want %q", got, payload)
	}
}

func TestParseLoginCallbackPreservesBase64Plus(t *testing.T) {
	t.Parallel()

	got, err := parseLoginCallback(fmt.Sprintf("http://localhost:%d/done?p=+w==", ServerPort))
	if err != nil {
		t.Fatalf("parse callback: %v", err)
	}

	if len(got) != 1 || got[0] != 0xfb {
		t.Fatalf("payload = %v, want [251]", got)
	}
}

func TestParseLoginCallbackRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"wrong scheme":  fmt.Sprintf("https://localhost:%d/done?p=e30=", ServerPort),
		"wrong host":    "http://example.com/done?p=e30=",
		"wrong port":    "http://localhost:9999/done?p=e30=",
		"wrong path":    fmt.Sprintf("http://localhost:%d/other?p=e30=", ServerPort),
		"missing value": fmt.Sprintf("http://localhost:%d/done", ServerPort),
		"extra query":   fmt.Sprintf("http://localhost:%d/done?p=e30=&other=value", ServerPort),
		"invalid base64": fmt.Sprintf(
			"http://localhost:%d/done?p=not-base64",
			ServerPort,
		),
	}

	for name, callback := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := parseLoginCallback(callback)
			if err == nil {
				t.Fatal("expected callback to be rejected")
			}
		})
	}
}

func TestReadInputLineAcceptsLongCallback(t *testing.T) {
	t.Parallel()

	want := strings.Repeat("a", 32_000)

	got, err := readInputLine(strings.NewReader(want + "\n"))
	if err != nil {
		t.Fatalf("read input: %v", err)
	}

	if string(got) != want {
		t.Fatalf("input length = %d, want %d", len(got), len(want))
	}
}
