package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version
var CmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Get manager version",
	Long:  `Mainflux server health checkt.`,
	Run: func(cmdCobra *cobra.Command, args []string) {
		Version()
	},
}

// Version - server health check
func Version() {
	url := fmt.Sprintf("%s/version", serverAddr)
	FormatResLog(httpClient.Get(url))
}
