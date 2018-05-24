package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/mainflux/mainflux/clients"
	"github.com/spf13/cobra"
)

var commands = []cobra.Command{
	cobra.Command{
		Use:   "create",
		Short: "create device/<JSON_client> <user_auth_token>",
		Long:  `Create new client, generate his UUID and store it`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 2 {
				msg := args[0]
				token := args[1]
				CreateClient(msg, token)
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
	cobra.Command{
		Use:   "get",
		Short: "get <user_auth_token> or get <client_id> <user_auth_token>",
		Long:  `Get all clients or client by id`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 1 {
				GetClients(args[0])
			} else if len(args) == 2 {
				GetClient(args[0], args[1])
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
	cobra.Command{
		Use:   "delete",
		Short: "delete all/<client_id> <user_auth_token>",
		Long:  `Removes client from database`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 2 {
				if args[0] == "all" {
					DeleteAllClients(args[1])
				} else {
					DeleteClient(args[0], args[1])
				}
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
	cobra.Command{
		Use:   "update",
		Short: "update <client_id> <JSON_string> <user_auth_token>",
		Long:  `Update client record`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) == 3 {
				UpdateClient(args[0], args[1], args[2])
			} else {
				LogUsage(cmdCobra.Short)
			}
		},
	},
	cobra.Command{
		Use:   "connect",
		Short: "connect <client_id> <channel_id> <user_auth_token>",
		Long:  `Connect client to the channel`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) != 3 {
				LogUsage(cmdCobra.Short)
			}
			ConnectClient(args[0], args[1], args[2])
		},
	},
	cobra.Command{
		Use:   "disconnect",
		Short: "disconnect <client_id> <channel_id> <user_auth_token>",
		Long:  `Disconnect client to the channel`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			if len(args) != 3 {
				LogUsage(cmdCobra.Short)
			}
			DisconnectClient(args[0], args[1], args[2])
		},
	},
}

// New does what godoc says...
func NewCmdClients() *cobra.Command {
	// package root
	cmd := cobra.Command{
		Use:   "clients",
		Short: "clients <options>",
		Long:  `Clients handling: create, delete or update clients.`,
		Run: func(cmdCobra *cobra.Command, args []string) {
			LogUsage(cmdCobra.Short)
		},
	}

	for i, _ := range commands {
		cmd.AddCommand(&commands[i])
	}

	return &cmd
}

var cliEndPoint = "clients"

// CreateClient - creates new client and generates client UUID
func CreateClient(msg, token string) {
	url := fmt.Sprintf("%s/%s", serverAddr, cliEndPoint)
	req, err := http.NewRequest("POST", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// GetClients - gets all clients
func GetClients(token string) {
	url := fmt.Sprintf("%s/%s?offset=%s&limit=%s",
		serverAddr, cliEndPoint, strconv.Itoa(Offset), strconv.Itoa(Limit))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// GetClient - gets client by ID
func GetClient(id, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, cliEndPoint, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// UpdateClient - updates client by ID
func UpdateClient(id, msg, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, cliEndPoint, id)
	req, err := http.NewRequest("PUT", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DeleteClient - removes client
func DeleteClient(id, token string) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, cliEndPoint, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DeleteAllClients - removes all clients
func DeleteAllClients(token string) {
	url := fmt.Sprintf("%s/%s", serverAddr, cliEndPoint)
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
		Clients []clients.Client `json:"clients,omitempty"`
	}
	json.Unmarshal([]byte(body), &list)

	for i := 0; i < len(list.Clients); i++ {
		DeleteClient(list.Clients[i].ID, token)
	}
}

// ConnectClient - connect client to a channel
func ConnectClient(cliId, chanId, token string) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, chanEndPoint,
		chanId, cliEndPoint, cliId)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}

// DisconnectClient - connect client to a channel
func DisconnectClient(cliId, chanId, token string) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, chanEndPoint,
		chanId, cliEndPoint, cliId)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	GetReqResp(req, token)
}
