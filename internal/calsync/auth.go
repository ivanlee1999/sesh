package calsync

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

const callbackPort = 19876

// RunAuthFlow performs the OAuth2 authorization code flow using a local HTTP server.
// It prints the auth URL, waits for the callback, exchanges the code, and returns the token.
func RunAuthFlow(oauthCfg *oauth2.Config) (*oauth2.Token, error) {
	return runAuthFlowInternal(oauthCfg, false)
}

// RunAuthFlowQuiet performs the OAuth2 flow without printing to stdout.
// Use this when running inside a TUI to avoid corrupting the alternate screen.
func RunAuthFlowQuiet(oauthCfg *oauth2.Config) (*oauth2.Token, error) {
	return runAuthFlowInternal(oauthCfg, true)
}

func runAuthFlowInternal(oauthCfg *oauth2.Config, quiet bool) (*oauth2.Token, error) {
	// Set redirect URL
	oauthCfg.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", callbackPort)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "no code in callback"
			}
			fmt.Fprintf(w, "Authorization failed: %s\nYou can close this window.", errMsg)
			errCh <- fmt.Errorf("auth callback error: %s", errMsg)
			return
		}
		fmt.Fprint(w, "Authorization successful! You can close this window and return to the terminal.")
		codeCh <- code
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", callbackPort))
	if err != nil {
		return nil, fmt.Errorf("start callback server on port %d: %w", callbackPort, err)
	}

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	// Generate auth URL
	authURL := oauthCfg.AuthCodeURL("sesh-auth", oauth2.AccessTypeOffline)

	if !quiet {
		fmt.Println("\nOpen this URL in your browser to authorize sesh:")
		fmt.Printf("  %s\n\n", authURL)
		fmt.Println("Waiting for authorization...")
	}

	// Try to open browser automatically
	openBrowser(authURL)

	// Wait for code or timeout
	select {
	case code := <-codeCh:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		token, err := oauthCfg.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("exchange code for token: %w", err)
		}
		return token, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("authorization timed out after 2 minutes")
	}
}

// openBrowser tries to open a URL in the default browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	_ = cmd.Start()
}
