package main

import (
	"log"

	bf "github.com/mainflux/mainflux/bashflux"
	"github.com/spf13/cobra"
)

func main() {

	conf := struct {
		HTTPHost string
		HTTPPort int
	}{
		"localhost",
		0,
	}

	// Root
	var rootCmd = &cobra.Command{
		Use: "bashflux",
		PersistentPreRun: func(cmdCobra *cobra.Command, args []string) {
			// Set HTTP server address
			bf.SetServerAddr(conf.HTTPHost, conf.HTTPPort)
		},
	}

	// API commands
	cmdUsers := bf.NewCmdUsers()
	cmdThings := bf.NewCmdThings()
	cmdChannels := bf.NewCmdChannels()

	// Root Commands
	rootCmd.AddCommand(cmdUsers)
	rootCmd.AddCommand(bf.CmdVersion)
	rootCmd.AddCommand(cmdThings)
	rootCmd.AddCommand(cmdChannels)
	rootCmd.AddCommand(bf.CmdMessages)

	// Messages
	bf.CmdMessages.AddCommand(bf.CmdSendMessage)

	// Root Flags
	rootCmd.PersistentFlags().StringVarP(
		&conf.HTTPHost, "host", "m", conf.HTTPHost, "HTTP Host address")
	rootCmd.PersistentFlags().IntVarP(
		&conf.HTTPPort, "port", "p", conf.HTTPPort, "HTTP Host Port")

	// Client and Channels Flags
	rootCmd.PersistentFlags().IntVarP(
		&bf.Limit, "limit", "l", 100, "limit query parameter")
	rootCmd.PersistentFlags().IntVarP(
		&bf.Offset, "offset", "o", 0, "offset query parameter")

	// Set TLS certificates
	bf.SetCerts()

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
