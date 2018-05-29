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

var thingsEP = "things"

var cmdThings = []cobra.Command{
	cobra.Command{
		Use:   "create",
		Short: "create device/<JSON_thing> <user_auth_token>",
		Long:  `Create new thing, generate his UUID and store it`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				LogUsage(cmd.Short)
				return
			}
			CreateThing(args[0], args[1])
		},
	},
	cobra.Command{
		Use:   "get",
		Short: "get all/<thing_id> <user_auth_token>",
		Long:  `Get all thingss or thing by id`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				LogUsage(cmd.Short)
				return
			}
			if args[0] == "all" {
				GetThings(args[1])
				return
			}
			GetThing(args[0], args[1])
		},
	},
	cobra.Command{
		Use:   "delete",
		Short: "delete all/<thing_id> <user_auth_token>",
		Long:  `Removes thing from database`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				LogUsage(cmd.Short)
				return
			}
			if args[0] == "all" {
				DeleteAllThings(args[1])
				return
			}
			DeleteThing(args[0], args[1])
		},
	},
	cobra.Command{
		Use:   "update",
		Short: "update <thing_id> <JSON_string> <user_auth_token>",
		Long:  `Update thing record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				LogUsage(cmd.Short)
				return
			}
			UpdateThing(args[0], args[1], args[2])
		},
	},
	cobra.Command{
		Use:   "connect",
		Short: "connect <thing_id> <channel_id> <user_auth_token>",
		Long:  `Connect thing to the channel`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				LogUsage(cmd.Short)
				return
			}
			ConnectThing(args[0], args[1], args[2])
		},
	},
	cobra.Command{
		Use:   "disconnect",
		Short: "disconnect <thing_id> <channel_id> <user_auth_token>",
		Long:  `Disconnect thing to the channel`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				LogUsage(cmd.Short)
				return
			}
			DisconnectThing(args[0], args[1], args[2])
		},
	},
}

func NewThingsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "things",
		Short: "things <options>",
		Long:  `Things handling: create, delete or update things.`,
		Run: func(cmd *cobra.Command, args []string) {
			LogUsage(cmd.Short)
		},
	}

	for i, _ := range cmdThings {
		cmd.AddCommand(&cmdThings[i])
	}

	return &cmd
}

// CreateThing - creates new thing and generates thing UUID
func CreateThing(msg, token string) {
	url := fmt.Sprintf("%s/%s", serverAddr, thingsEP)
	req, err := http.NewRequest("POST", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// GetThings - gets all things
func GetThings(token string) {
	url := fmt.Sprintf("%s/%s?offset=%s&limit=%s",
		serverAddr, thingsEP, strconv.Itoa(Offset), strconv.Itoa(Limit))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// GetThing - gets thing by ID
func GetThing(id, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEP, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// UpdateThing - updates thing by ID
func UpdateThing(id, msg, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEP, id)
	req, err := http.NewRequest("PUT", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DeleteThing - removes thing
func DeleteThing(id, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEP, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DeleteAllThings - removes all things
func DeleteAllThings(token string) {
	url := fmt.Sprintf("%s/%s", serverAddr, thingsEP)
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
		Things []things.Thing `json:"things,omitempty"`
	}
	json.Unmarshal([]byte(body), &list)

	for i := 0; i < len(list.Things); i++ {
		DeleteThing(string(list.Things[i].ID), token)
	}
}

// ConnectThing - connect thing to a channel
func ConnectThing(cliId, chanId, token string) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, chanEndPoint,
		chanId, thingsEP, cliId)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DisconnectThing - connect thing to a channel
func DisconnectThing(cliId, chanId, token string) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, chanEndPoint,
		chanId, thingsEP, cliId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}
