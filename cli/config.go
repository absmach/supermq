// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

type url struct {
	ThingsURL      string `toml:"ThingsURL"`
	UsersURL       string `toml:"UsersURL"`
	ReaderURL      string `toml:"ReaderURL"`
	HTTPAdapterURL string `toml:"HTTPAdapterURL"`
	BootstrapURL   string `toml:"BootstrapURL"`
	CertsURL       string `toml:"CertsURL"`
}

type filter struct {
	Offset          string `toml:"Offset"`
	Limit           string `toml:"Limit"`
	RawOutput       string `toml:"RawOutput"`
	Name            string `toml:"Name"`
	Contact         string `toml:"Contact"`
	Email           string `toml:"Email"`
	Metadata        string `toml:"Metadata"`
	MsgContentType  string `toml:"MsgContentType"`
	TLSVerification string `toml:"TLSVerification"`
}

type channel struct {
	Status string `toml:"Status"`
	State  string `toml:"State"`
	Topic  string `toml:"Topic"`
}

type Config struct {
	URL     url     `toml:"url"`
	Filter  filter  `toml:"filter"`
	Channel channel `toml:"channel"`
}

// read - retrieve config from a file.
func read(file string) (Config, error) {
	c := Config{}
	data, err := os.Open(file)
	if err != nil {
		return c, errors.New(fmt.Sprintf("failed to read config file: %s", err))
	}
	defer data.Close()

	buf, err := ioutil.ReadAll(data)
	if err != nil {
		return c, errors.New(fmt.Sprintf("failed to read config file: %s", err))
	}

	if err := toml.Unmarshal(buf, &c); err != nil {
		return Config{}, errors.New(fmt.Sprintf("failed to unmarshal config TOML: %s", err))
	}

	return c, nil
}

func ParseConfig() error {
	if ConfigPath == "" {
		// No config file
		return nil
	}

	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		errConfigNotFound := errors.Wrap(errors.New("config file was not found"), err)
		logError(errConfigNotFound)
		return nil
	}

	config, err := read(ConfigPath)
	if err != nil {
		return err
	}

	if config.Filter.Offset != "" {
		offset, err := strconv.ParseUint(config.Filter.Offset, 10, 64)
		if err != nil {
			fmt.Println("Error converting Offset to uint64:", err)
			return
		}
		Offset = offset
	}

	if config.Filter.Limit != "" {
		limit, err := strconv.ParseUint(config.Filter.Limit, 10, 64)
		if err != nil {
			fmt.Println("Error converting Offset to uint64:", err)
			return
		}
		Limit = limit
	}

	if config.Filter.Name != "" {
		Name = config.Filter.Name
	}

	if config.Filter.RawOutput != "" {
		rawOutput, err := strconv.ParseBool(config.Filter.RawOutput)

		if err != nil {
			fmt.Println("Error converting string to bool:", err)
		}

		RawOutput = rawOutput
	}
	return nil
}

func NewConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "CLI local config",
		Long:  "Local param storage to prevent repetitive passing of keys",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Configuring CLI...")
			var key, value string

			// Prompt the user for inputs
			fmt.Print("Enter key: ")
			fmt.Scan(&key)

			fmt.Print("Enter value: ")
			fmt.Scan(&value)

			setConfigValue(key, value)
			fmt.Println("Configuration complete")
		},
	}
}

func setConfigValue(key string, value string) {
	// Read the existing configuration file
	configPath := ConfigPath
	config, err := read(configPath)
	if err != nil {
		fmt.Println("Error reading the existing configuration:", err)
		return
	}

	// Update the specific key in the struct
	switch key {
	case "ThingsURL":
		config.URL.ThingsURL = value
	case "UsersURL":
		config.URL.UsersURL = value
	case "ReaderURL":
		config.URL.ReaderURL = value
	case "HTTPAdapterURL":
		config.URL.HTTPAdapterURL = value
	case "BootstrapURL":
		config.URL.BootstrapURL = value
	case "CertsURL":
		config.URL.CertsURL = value
	case "Offset":
		config.Filter.Offset = value
	case "Limit":
		config.Filter.Limit = value
	case "Name":
		config.Filter.Name = value
	case "RawOutput":
		config.Filter.RawOutput = value
	case "Status":
		config.Channel.Status = value
	case "State":
		config.Channel.State = value
	case "Topic":
		config.Channel.Topic = value
	default:
		fmt.Println("Unknown key:", key)
		return
	}

	// Marshal the updated struct back into TOML format
	buf, err := toml.Marshal(config)
	if err != nil {
		fmt.Println("Error marshaling the configuration:", err)
		return
	}

	// Write the updated configuration to the TOML file
	err = os.WriteFile(configPath, buf, 0644)
	if err != nil {
		fmt.Println("Error writing the updated configuration to file:", err)
		return
	}

}
