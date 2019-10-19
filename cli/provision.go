// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"

	"github.com/docker/docker/pkg/namesgenerator"
	mfxsdk "github.com/mainflux/mainflux/sdk/go"
	"github.com/spf13/cobra"
)

var errMalformedCSV = errors.New("malformed CSV")

func createThing(name, token string) (mfxsdk.Thing, error) {
	id, err := sdk.CreateThing(mfxsdk.Thing{Name: name}, token)
	if err != nil {
		return mfxsdk.Thing{}, err
	}

	t, err := sdk.Thing(id, token)
	if err != nil {
		return mfxsdk.Thing{}, err
	}

	m := mfxsdk.Thing{
		ID:   id,
		Name: name,
		Key:  t.Key,
	}

	return m, nil
}

func createChannel(name, token string) (mfxsdk.Channel, error) {
	id, err := sdk.CreateChannel(mfxsdk.Channel{Name: name}, token)
	if err != nil {
		return mfxsdk.Channel{}, nil
	}

	c := mfxsdk.Channel{
		ID:   id,
		Name: name,
	}

	return c, nil
}

var cmdProvision = []cobra.Command{
	cobra.Command{
		Use:   "things",
		Short: "things <things_csv> <user_token>",
		Long:  `Bulk create things`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}

			things, err := sdk.BulkCreateThings(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(things)
		},
	},
	cobra.Command{
		Use:   "channels",
		Short: "channels <channels_csv> <user_token>",
		Long:  `Bulk create channels`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}

			channels, err := sdk.BulkCreateChannels(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(channels)
		},
	},
	cobra.Command{
		Use:   "test",
		Short: "test",
		Long: `Provisions test setup: one test user, two things and two channels. \
						Connect both things to one of the channels, \
						and only on thing to other channel.`,
		Run: func(cmd *cobra.Command, args []string) {
			numThings := 2
			numChan := 2
			things := []mfxsdk.Thing{}
			channels := []mfxsdk.Channel{}

			if len(args) != 0 {
				logUsage(cmd.Short)
				return
			}

			un := fmt.Sprintf("%s@email.com", namesgenerator.GetRandomName(0))
			// Create test user
			user := mfxsdk.User{
				Email:    un,
				Password: "123",
			}
			if err := sdk.CreateUser(user); err != nil {
				logError(err)
				return
			}

			ut, err := sdk.CreateToken(user)
			if err != nil {
				logError(err)
				return
			}

			// Create things
			for i := 0; i < numThings; i++ {
				n := fmt.Sprintf("d%d", i)

				m, err := createThing(n, ut)
				if err != nil {
					logError(err)
					return
				}

				things = append(things, m)
			}
			// Create channels
			for i := 0; i < numChan; i++ {
				n := fmt.Sprintf("c%d", i)
				c, err := createChannel(n, ut)
				if err != nil {
					logError(err)
					return
				}

				channels = append(channels, c)
			}

			// Connect things to channels - first thing to both channels, second only to first
			for i := 0; i < numThings; i++ {
				if err := sdk.ConnectThing(things[i].ID, channels[i].ID, ut); err != nil {
					logError(err)
					return
				}

				if i%2 == 0 {
					if i+1 >= len(channels) {
						break
					}
					if err := sdk.ConnectThing(things[i].ID, channels[i+1].ID, ut); err != nil {
						logError(err)
						return
					}
				}
			}

			logJSON(user, ut, things, channels)
		},
	},
}

// NewProvisionCmd returns provision command.
func NewProvisionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "provision",
		Short: "Provision things and channels from a config file",
		Long:  `Provision things and channels: use json or csv file to bulk provision things and channels`,
	}

	for i := range cmdProvision {
		cmd.AddCommand(&cmdProvision[i])
	}

	return &cmd
}
