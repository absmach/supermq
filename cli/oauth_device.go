// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	pollInterval = 3 * time.Second
	pollTimeout  = 10 * time.Minute
)

func performOAuthDeviceLogin(cmd *cobra.Command, provider string) error {
	ctx := cmd.Context()

	// Step 1: Get device code
	deviceCode, err := sdk.OAuthDeviceCode(ctx, provider)
	if err != nil {
		return fmt.Errorf("failed to get device code: %w", err)
	}

	// Step 2: Display instructions to user
	printDeviceInstructions(deviceCode.VerificationURI, deviceCode.UserCode)

	// Step 3: Poll for authorization
	token, pollErr := pollForAuthorization(ctx, provider, deviceCode.DeviceCode, deviceCode.Interval)
	if pollErr != nil {
		return fmt.Errorf("authorization failed: %w", pollErr)
	}

	// Step 4: Display success message
	logJSONCmd(*cmd, token)
	successMsg := color.New(color.FgGreen, color.Bold).SprintFunc()
	fmt.Printf("\n%s\n", successMsg("✓ Authentication successful!"))
	fmt.Println("You can now use the access_token for API requests.")

	return nil
}

func printDeviceInstructions(verificationURI, userCode string) {
	fmt.Println()
	fmt.Println(color.New(color.FgCyan, color.Bold).Sprint("=== OAuth Device Authorization ==="))
	fmt.Println()
	fmt.Println(color.New(color.FgYellow).Sprint("Please complete authentication in your browser:"))
	fmt.Println()
	fmt.Printf("  1. Visit: %s\n", color.New(color.FgBlue, color.Underline).Sprint(verificationURI))
	fmt.Printf("  2. Enter code: %s\n", color.New(color.FgGreen, color.Bold).Sprint(userCode))
	fmt.Println()
	fmt.Println(color.New(color.FgWhite).Sprint("Waiting for authorization..."))
	fmt.Println()
}

func pollForAuthorization(ctx context.Context, provider, deviceCode string, interval int) (interface{}, error) {
	pollDuration := time.Duration(interval) * time.Second
	if pollDuration < pollInterval {
		pollDuration = pollInterval
	}

	ticker := time.NewTicker(pollDuration)
	defer ticker.Stop()

	timeout := time.After(pollTimeout)
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIdx := 0

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("authentication timeout after %v", pollTimeout)

		case <-ticker.C:
			// Show spinner
			fmt.Printf("\r%s Polling for authorization...", color.CyanString(spinner[spinnerIdx]))
			spinnerIdx = (spinnerIdx + 1) % len(spinner)

			token, err := sdk.OAuthDeviceToken(ctx, provider, deviceCode)
			if err != nil {
				errMsg := err.Error()
				// Check if it's a pending error (expected during polling)
				if strings.Contains(errMsg, "authorization pending") || strings.Contains(errMsg, "slow down") {
					continue
				}
				// Any other error is a real failure
				return nil, fmt.Errorf("failed to get token: %w", err)
			}

			// Clear the spinner line
			fmt.Print("\r" + string(make([]byte, 50)) + "\r")
			return token, nil
		}
	}
}
