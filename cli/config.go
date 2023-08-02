// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/mainflux/mainflux/pkg/errors"
	mfxsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

type remotes struct {
	ThingsURL       string `toml:"things_url"`
	UsersURL        string `toml:"users_url"`
	ReaderURL       string `toml:"reader_url"`
	HTTPAdapterURL  string `toml:"http_adapter_url"`
	BootstrapURL    string `toml:"bootstrap_url"`
	CertsURL        string `toml:"certs_url"`
	MsgContentType  string `toml:"msg_content_type"`
	TLSVerification bool   `toml:"tls_verification"`
}

type filter struct {
	Offset    string `toml:"offset"`
	Limit     string `toml:"limit"`
	RawOutput string `toml:"raw_output"`
	Name      string `toml:"name"`
	Contact   string `toml:"contact"`
	Email     string `toml:"email"`
	Metadata  string `toml:"metadata"`
}

type channel struct {
	Status string `toml:"status"`
	State  string `toml:"state"`
	Topic  string `toml:"topic"`
}

type config struct {
	Remotes remotes `toml:"remotes"`
	Filter  filter  `toml:"filter"`
	Channel channel `toml:"channel"`
}

var (
	errReadFail            = errors.New("failed to read config file")
	errUnmarshalFail       = errors.New("failed to Unmarshall config TOML")
	errConfigNotFound      = errors.New("config file was not found")
	errUintConv            = errors.New("error converting filter to Uint64")
	errBoolConv            = errors.New("error converting string to bool")
	errUseExistConf        = errors.New("error using the existing configuration")
	errNoKey               = errors.New("no such key")
	errInvalidInt          = errors.New("error: invalid integer value for key")
	errInvalidBool         = errors.New("error: invalid boolean value for key")
	errUnsupportedKeyValue = errors.New("error: unsupported data type for key")
	errMarshal             = errors.New("error marshaling the configuration")
	errWritingConfig       = errors.New("error writing the updated config to file")
	errInvalidURL          = errors.New("invalid url")
	ConfigFile             = ""
)

// func read(file string) (config, error) {
// 	c := config{}
// 	data, err := os.Open(file)
// 	if err != nil {
// 		return c, errors.Wrap(errReadFail, err)
// 	}
// 	defer data.Close()

// 	buf, err := io.ReadAll(data)
// 	if err != nil {
// 		return c, errors.Wrap(errReadFail, err)
// 	}

// 	if err := toml.Unmarshal(buf, &c); err != nil {
// 		return config{}, errors.Wrap(errUnmarshalFail, err)
// 	}

// 	return c, nil
// }

func read(file string) (config, error) {
	c := config{}
	data, err := os.Open(file)
	if err != nil {
		return c, errors.Wrap(errReadFail, err)
	}
	defer data.Close()

	buf, err := io.ReadAll(data)
	if err != nil {
		return c, errors.Wrap(errReadFail, err)
	}

	if err := toml.Unmarshal(buf, &c); err != nil {
		return config{}, errors.Wrap(errUnmarshalFail, err)
	}

	return c, nil
}

func ParseConfig() error {
	if ConfigPath == "" {
		// No config file
		return nil
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		errConfigNotFound := errors.Wrap(errConfigNotFound, err)
		logError(errConfigNotFound)
		return nil
	}

	config, err := read(configFile)
	if err != nil {
		return err
	}
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		errConfigNotFound := errors.Wrap(errConfigNotFound, err)
		logError(errConfigNotFound)
		return sdkConf
	}

	if config.Filter.Offset != "" {
		offset, err := strconv.ParseUint(config.Filter.Offset, 10, 64)
		if err != nil {
			logError(errors.Wrap(errUintConv, err))
			return sdkConf
		}
		Offset = offset
	}

	if config.Filter.Limit != "" {
		limit, err := strconv.ParseUint(config.Filter.Limit, 10, 64)
		if err != nil {
			logError(errors.Wrap(errUintConv, err))
			return sdkConf
		}
		Limit = limit
	}

	if config.Filter.Name != "" {
		Name = config.Filter.Name
	}

	if config.Filter.RawOutput != "" {
		rawOutput, err := strconv.ParseBool(config.Filter.RawOutput)
		if err != nil {
			logError(errors.Wrap(errBoolConv, err))
		}

		RawOutput = rawOutput
	}
	return nil
}

// New config command to store params to local TOML file
// func NewConfigCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "config <key> <value>",
// 		Short: "CLI local config",
// 		Long:  "Local param storage to prevent repetitive passing of keys",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			if len(args) != 2 {
// 				logUsage(cmd.Use)
// 				return
// 			}

// 			key := args[0]
// 			value := args[1]

// 			setConfigValue(key, value)
// 			logOK()
// 		},
// 	}
// }

// New config command to store params to local TOML file
func NewConfigCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "config <key> <value>",
		Short: "CLI local config",
		Long:  "Local param storage to prevent repetitive passing of keys",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			key := args[0]
			value := args[1]

			setConfigValue(key, value)
			logOK()
		},
	}

	cmd.Flags().StringVarP(&ConfigFile, "config", "c", "", "Config file path")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			logUsage(cmd.Use)
			return
		}

		key := args[0]
		value := args[1]

		// ParseConfig(sdkConf, configFile) // Use the provided configFile
		setConfigValue(key, value)
		logOK()
	}

	return cmd
}

func setConfigValue(key string, value string) {
	config, err := read(ConfigPath)
	if err != nil {
		logError(errors.Wrap(errUseExistConf, err))
		return
	}

	if isURLKey(key) {
		if !isValidURL(value) {
			logError(errInvalidURL)
			return
		}
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
		"metadata":         &config.Filter.Metadata,
		"tls_verification": &config.Remotes.TLSVerification,
		"msg_content_type": &config.Remotes.MsgContentType,
	}

	fieldPtr, found := configKeyToField[key]
	if !found {
		logError(errNoKey)
		return
	}

	fieldValue := reflect.ValueOf(fieldPtr).Elem()

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			logError(errors.Wrap(errInvalidInt, err))
			return
		}
		fieldValue.SetInt(int64(intValue))
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			logError(errors.Wrap(errInvalidBool, err))
			return
		}
		fieldValue.SetBool(boolValue)
	default:
		logError(errors.Wrap(errUnsupportedKeyValue, err))
		return
	}

	buf, err := toml.Marshal(config)
	if err != nil {
		logError(errors.Wrap(errMarshal, err))
		return
	}

	err = os.WriteFile(ConfigPath, buf, 0644)
	if err != nil {
		logError(errors.Wrap(errWritingConfig, err))
		return
	}
}

func isValidURL(inputURL string) bool {
	u, err := url.Parse(inputURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return strings.HasPrefix(u.Scheme, "http") || strings.HasPrefix(u.Scheme, "https")
}

func isURLKey(key string) bool {
	urlKeys := []string{
		"things_url",
		"users_url",
		"reader_url",
		"http_adapter_url",
		"bootstrap_url",
		"certs_url",
	}

	for _, urlKey := range urlKeys {
		if key == urlKey {
			return true
		}
	}
	return false
}
