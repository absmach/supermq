// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

const (
	callbackPath    = "/callback"
	localServerPort = "9090"
	callbackTimeout = 5 * time.Minute
	shutdownTimeout = 5 * time.Second
)

type oauthCallbackResult struct {
	code  string
	state string
	err   error
}

type browserOpener interface {
	Open(url string) error
}

type defaultBrowserOpener struct{}

func (defaultBrowserOpener) Open(url string) error {
	return openBrowser(url)
}

type callbackServer struct {
	listener net.Listener
	server   *http.Server
}

func newCallbackServer(resultChan chan<- oauthCallbackResult) (*callbackServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+localServerPort)
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}

	mux := http.NewServeMux()
	server := &http.Server{Handler: mux}

	var once sync.Once
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(w, r, resultChan, &once)
	})

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			once.Do(func() {
				resultChan <- oauthCallbackResult{err: fmt.Errorf("server error: %w", err)}
			})
		}
	}()

	return &callbackServer{
		listener: listener,
		server:   server,
	}, nil
}

func (cs *callbackServer) Shutdown(cmd *cobra.Command) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := cs.server.Shutdown(ctx); err != nil {
		logErrorCmd(*cmd, err)
	}
}

func performOAuthLogin(cmd *cobra.Command, provider string) error {
	return performOAuthLoginWithBrowser(cmd, provider, defaultBrowserOpener{})
}

func performOAuthLoginWithBrowser(cmd *cobra.Command, provider string, browser browserOpener) error {
	ctx := cmd.Context()
	callbackChan := make(chan oauthCallbackResult, 1)

	server, err := newCallbackServer(callbackChan)
	if err != nil {
		return err
	}
	defer server.Shutdown(cmd)

	callbackURL := fmt.Sprintf("http://127.0.0.1:%s%s", localServerPort, callbackPath)
	authURL, state, err := sdk.OAuthAuthorizationURL(ctx, provider, callbackURL)
	if err != nil {
		return fmt.Errorf("failed to get authorization URL: %w", err)
	}

	printAuthInstructions(authURL)
	if err := browser.Open(authURL); err != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", err)
	}

	fmt.Println("Waiting for authentication callback...")

	result, err := waitForCallback(callbackChan)
	if err != nil {
		return err
	}

	fmt.Println("Exchanging authorization code for tokens...")
	token, err := sdk.OAuthCallback(ctx, provider, result.code, state, callbackURL)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	logJSONCmd(*cmd, token)
	fmt.Println("\nAuthentication successful! You can now use the access_token for API requests.")

	return nil
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request, resultChan chan<- oauthCallbackResult, once *sync.Once) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errParam := r.URL.Query().Get("error")

	w.Header().Set("Content-Type", "text/html")

	if errParam != "" {
		html := strings.Replace(errorHTML, "{{ERROR_MESSAGE}}", errParam, 1)
		fmt.Fprint(w, html)
		once.Do(func() {
			resultChan <- oauthCallbackResult{err: fmt.Errorf("oauth error: %s", errParam)}
		})
		return
	}

	if code == "" {
		html := strings.Replace(errorHTML, "{{ERROR_MESSAGE}}", "missing authorization code", 1)
		fmt.Fprint(w, html)
		once.Do(func() {
			resultChan <- oauthCallbackResult{err: fmt.Errorf("missing authorization code")}
		})
		return
	}

	fmt.Fprint(w, successHTML)
	once.Do(func() {
		resultChan <- oauthCallbackResult{code: code, state: state}
	})
}

func printAuthInstructions(authURL string) {
	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If the browser doesn't open automatically, please visit:\n%s\n\n", authURL)
}

func waitForCallback(callbackChan <-chan oauthCallbackResult) (oauthCallbackResult, error) {
	select {
	case result := <-callbackChan:
		if result.err != nil {
			return oauthCallbackResult{}, fmt.Errorf("callback error: %w", result.err)
		}
		return result, nil
	case <-time.After(callbackTimeout):
		return oauthCallbackResult{}, fmt.Errorf("authentication timeout after %v", callbackTimeout)
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
