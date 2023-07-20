// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

type remotes struct {
	ThingsURL      string `toml:"things_url"`
	UsersURL       string `toml:"users_url"`
	ReaderURL      string `toml:"reader_url"`
	HTTPAdapterURL string `toml:"http_adapter_url"`
	BootstrapURL   string `toml:"bootstrap_url"`
	CertsURL       string `toml:"certs_url"`
}

type filter struct {
	Offset          string `toml:"offset"`
	Limit           string `toml:"limit"`
	RawOutput       string `toml:"raw_output"`
	Name            string `toml:"name"`
	Contact         string `toml:"contact"`
	Email           string `toml:"email"`
	Metadata        string `toml:"metadata"`
	MsgContentType  string `toml:"msg_content_type"`
	TLSVerification string `toml:"tls_verification"`
}

type channel struct {
	Status string `toml:"status"`
	State  string `toml:"state"`
	Topic  string `toml:"topic"`
}

type Config struct {
	Remotes remotes `toml:"remotes"`
	Filter  filter  `toml:"filter"`
	Channel channel `toml:"channel"`
}

// read - retrieve config from a file.
func read(file string) (Config, error) {
	c := Config{}
	data, err := os.Open(file)
	if err != nil {
		return c, err
	}
	defer data.Close()

	buf, err := io.ReadAll(data)
	if err != nil {
		return c, err
	}

	if err := toml.Unmarshal(buf, &c); err != nil {
		return Config{}, err
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
			log.Println("Error converting Offset to uint64:", err)
			return
		}
		Offset = offset
	}

	if config.Filter.Limit != "" {
		limit, err := strconv.ParseUint(config.Filter.Limit, 10, 64)
		if err != nil {
			log.Println("Error converting Offset to uint64:", err)
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
			log.Println("Error converting string to bool:", err)
		}

		RawOutput = rawOutput
	}
	return nil
}

func NewConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config <key> <value>",
		Short: "CLI local config",
		Long:  "Local param storage to prevent repetitive passing of keys",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Configuring CLI...")

			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			key := args[0]
			value := args[1]
			// Prompt the user for inputs
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
		log.Println("Error reading the existing configuration:", err)
		return
	}

	var configKeyToField = map[string]interface{}{
		"things_url":       &config.Remotes.ThingsURL,
		"users_url":        &config.Remotes.UsersURL,
		"reader_url":       &config.Remotes.ReaderURL,
		"http_adapter_url": &config.Remotes.HTTPAdapterURL,
		"bootstrap_url":    &config.Remotes.BootstrapURL,
		"certs_url":        &config.Remotes.CertsURL,
		"offset":           &config.Filter.Offset,
		"limit":            &config.Filter.Limit,
		"name":             &config.Filter.Name,
		"raw_output":       &config.Filter.RawOutput,
		"status":           &config.Channel.Status,
		"state":            &config.Channel.State,
		"topic":            &config.Channel.Topic,
	}

	fieldPtr, found := configKeyToField[key]
	if !found {
		log.Println("Unknown key:", key)
		return
	}

	fieldValue := reflect.ValueOf(fieldPtr).Elem()

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			fmt.Println("Error: Invalid integer value for key", key)
			return
		}
		fieldValue.SetInt(int64(intValue))
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			fmt.Println("Error: Invalid boolean value for key", key)
			return
		}
		fieldValue.SetBool(boolValue)
	default:
		fmt.Println("Error: Unsupported data type for key", key)
		return
	}

	// Marshal the updated struct back into TOML format
	buf, err := toml.Marshal(config)
	if err != nil {
		log.Println("Error marshaling the configuration:", err)
		return
	}

	// Write the updated configuration to the TOML file
	err = os.WriteFile(configPath, buf, 0644)
	if err != nil {
		log.Println("Error writing the updated configuration to file:", err)
		return
	}

}
