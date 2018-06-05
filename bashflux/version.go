package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "Get version of Mainflux Things Service",
		Long:  `Mainflux server health checkt.`,
		Run: func(cmd *cobra.Command, args []string) {
			Version()
		},
	}

	return &cmd
}

// Version - server health check
func Version() {
	url := fmt.Sprintf("%s/things/version", serverAddr)
	FormatResLog(httpClient.Get(url))
}
