// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"

	mfxsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/spf13/cobra"
)

var cmdThings = []cobra.Command{
	{
		Use:   "create <JSON_thing> <user_auth_token>",
		Short: "Create thing",
		Long:  `Create new thing, generate his UUID and store it`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			var thing mfxsdk.Thing
			if err := json.Unmarshal([]byte(args[0]), &thing); err != nil {
				logError(err)
				return
			}
			thing.Status = clients.EnabledStatus.String()
			thing, err := sdk.CreateThing(thing, args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "get [all | <thing_id>] <user_auth_token>",
		Short: "Get things",
		Long: `Get all things or get thing by id. Things can be filtered by name or metadata
		all - lists all things
		<thing_id> - shows thing with provided <thing_id>`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}
			metadata, err := convertMetadata(Metadata)
			if err != nil {
				logError(err)
				return
			}
			pageMetadata := mfxsdk.PageMetadata{
				Name:     "",
				Offset:   uint64(Offset),
				Limit:    uint64(Limit),
				Metadata: metadata,
			}
			if args[0] == "all" {
				l, err := sdk.Things(pageMetadata, args[1])
				if err != nil {
					logError(err)
					return
				}
				logJSON(l)
				return
			}
			t, err := sdk.Thing(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(t)
		},
	},
	{
		Use:   "identify <thing_key>",
		Short: "Identify thing",
		Long:  "Validates thing's key and returns its ID",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				logUsage(cmd.Use)
				return
			}

			i, err := sdk.IdentifyThing(args[0])
			if err != nil {
				logError(err)
				return
			}

			logJSON(i)
		},
	},
	{
		Use:   "update <JSON_string> <user_auth_token>",
		Short: "Update thing",
		Long:  `Update thing record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			var thing mfxsdk.Thing
			if err := json.Unmarshal([]byte(args[0]), &thing); err != nil {
				logError(err)
				return
			}

			thing, err := sdk.UpdateThing(thing, args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "updatetags <thing_id> <tags> <user_auth_token>",
		Short: "Update thing tags",
		Long:  `Update thing record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Use)
				return
			}

			var thing mfxsdk.Thing
			if err := json.Unmarshal([]byte(args[1]), &thing.Tags); err != nil {
				logError(err)
				return
			}
			thing.ID = args[0]
			thing, err := sdk.UpdateThingTags(thing, args[2])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "updateidentity <thing_id> <identity> <user_auth_token>",
		Short: "Update thing identity",
		Long:  `Update thing record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Use)
				return
			}

			var thing mfxsdk.Thing
			if err := json.Unmarshal([]byte(args[1]), &thing.Credentials.Identity); err != nil {
				logError(err)
				return
			}
			thing.ID = args[0]
			thing, err := sdk.UpdateThingIdentity(thing, args[2])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "updatesecret <thing_id> <secret> <user_auth_token>",
		Short: "Update thing tags",
		Long:  `Update thing record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Use)
				return
			}

			thing, err := sdk.UpdateThingSecret(args[0], args[1], args[2])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "updateowner <thing_id> <tags> <user_auth_token>",
		Short: "Update thing owner",
		Long:  `Update thing record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Use)
				return
			}

			var thing mfxsdk.Thing
			if err := json.Unmarshal([]byte(args[1]), &thing.Owner); err != nil {
				logError(err)
				return
			}
			thing.ID = args[0]
			thing, err := sdk.UpdateThingOwner(thing, args[2])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "enable <thing_id> <user_auth_token>",
		Short: "Change thing status to enabled",
		Long:  `Change thing status to enabled`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			thing, err := sdk.EnableThing(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "disable <thing_id> <user_auth_token>",
		Short: "Change thing status to disabled",
		Long:  `Change thing status to disabled`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			thing, err := sdk.DisableThing(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(thing)
		},
	},
	{
		Use:   "connect <thing_id> <channel_id> <user_auth_token>",
		Short: "Connect thing",
		Long:  `Connect thing to the channel`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Use)
				return
			}

			connIDs := mfxsdk.ConnectionIDs{
				ChannelIDs: []string{args[1]},
				ThingIDs:   []string{args[0]},
			}
			if err := sdk.Connect(connIDs, args[2]); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	},
	{
		Use:   "disconnect <thing_id> <channel_id> <user_auth_token>",
		Short: "Disconnect thing",
		Long:  `Disconnect thing to the channel`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Use)
				return
			}

			connIDs := mfxsdk.ConnectionIDs{
				ThingIDs:   []string{args[0]},
				ChannelIDs: []string{args[1]},
			}
			if err := sdk.Disconnect(connIDs, args[2]); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	},
	{
		Use:   "connections <thing_id> <user_auth_token>",
		Short: "Connected list",
		Long:  `List of Channels connected to Thing`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}
			pm := mfxsdk.PageMetadata{
				Offset:       uint64(Offset),
				Limit:        uint64(Limit),
				Disconnected: false,
			}
			cl, err := sdk.ChannelsByThing(args[0], pm, args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(cl)
		},
	},
}

// NewThingsCmd returns things command.
func NewThingsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "things [create | get | update | delete | connect | disconnect | connections | not-connected]",
		Short: "Things management",
		Long:  `Things management: create, get, update or delete Thing, connect or disconnect Thing from Channel and get the list of Channels connected or disconnected from a Thing`,
	}

	for i := range cmdThings {
		cmd.AddCommand(&cmdThings[i])
	}

	return &cmd
}
