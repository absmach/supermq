package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// New does what godoc says...
func NewVersionCmd() *cobra.Command {
	// package root
	cmd := cobra.Command{
		Use:   "version",
		Short: "Get manager version",
		Long:  `Mainflux server health checkt.`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			Version()
		},
	}

	return &cmd
}

// Version - server health check
func Version() {
	url := fmt.Sprintf("%s/version", serverAddr)
	FormatResLog(httpClient.Get(url))
}
