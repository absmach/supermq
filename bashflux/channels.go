package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/mainflux/mainflux/things"
	"github.com/spf13/cobra"
)

var chanEndPoint = "channels"

var cmdChannels = []cobra.Command{
	cobra.Command{
		Use:   "create",
		Short: "create <JSON_channel> <user_auth_token>",
		Long:  `Creates new channel and generates it's UUID`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 2 {
				msg := args[0]
				token := args[1]
				CreateChannel(msg, token)
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
	cobra.Command{
		Use:   "get",
		Short: "get <user_auth_token> or get <channel_id> <user_auth_token>",
		Long:  `Gets list of all channels or gets channel by id`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 1 {
				GetChannels(args[0])
			} else if len(args) == 2 {
				GetChannel(args[0], args[1])
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
	cobra.Command{
		Use:   "update",
		Short: "update <channel_id> <JSON_string> <user_auth_token>",
		Long:  `Updates channel record`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 3 {
				UpdateChannel(args[0], args[1], args[2])
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
	cobra.Command{
		Use:   "delete",
		Short: "delete <channel_id> <user_auth_token>",
		Long:  `Delete channel by ID`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 2 {
				if args[0] == "all" {
					DeleteAllChannels(args[1])
				} else {
					DeleteChannel(args[0], args[1])
				}
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
}

// New does what godoc says...
func NewChannelsCmd() *cobra.Command {
	// package root
	cmd := cobra.Command{
		Use:   "channels",
		Short: "Manipulation with channels",
		Long:  `Manipulation with channels: create, delete or update channels`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			LogUsage(cmdCobra.Short)
		},
	}

	for i, _ := range cmdChannels {
		cmd.AddCommand(&cmdChannels[i])
	}

	return &cmd
}

// CreateChannel - creates new channel and generates UUID
func CreateChannel(msg string, token string) {
	url := fmt.Sprintf("%s/%s", serverAddr, chanEndPoint)
	req, err := http.NewRequest("POST", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// GetChannels - gets all channels
func GetChannels(token string) {
	url := fmt.Sprintf("%s/%s?offset=%s&limit=%s",
		serverAddr, chanEndPoint, strconv.Itoa(Offset), strconv.Itoa(Limit))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// GetChannel - gets channel by ID
func GetChannel(id string, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, chanEndPoint, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// UpdateChannel - publishes SenML message on the channel
func UpdateChannel(id string, msg string, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, chanEndPoint, id)
	req, err := http.NewRequest("PUT", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DeleteChannel - removes channel
func DeleteChannel(id, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, chanEndPoint, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DeleteAllChannels - removes all channels
func DeleteAllChannels(token string) {
	url := fmt.Sprintf("%s/%s", serverAddr, chanEndPoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	req.Header.Set("Authorization", token)
	req.Header.Add("Content-Type", contentType)

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	var list struct {
		Channels []things.Channel `json:"channels,omitempty"`
	}
	json.Unmarshal([]byte(body), &list)

	for i := 0; i < len(list.Channels); i++ {
		DeleteChannel(string(list.Channels[i].ID), token)
	}
}
