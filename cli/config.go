// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"os"
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
	errURLParseFail        = errors.New("failed to parse url")
	defaultConfigPath      = "./config.toml"
)

// configFileContents contains the contents of the config file in hex.
const configFileContents = "0a5b6368616e6e656c5d0a20207374617465203d2022220a2020737461747573203d2022220a2020746f706963203d2022220a0a5b66696c7465725d0a2020636f6e74616374203d2022220a2020656d61696c203d2022220a20206c696d6974203d2022220a20206d65746164617461203d2022220a20206e616d65203d2022220a20206f6666736574203d2022220a20207261775f6f7574707574203d2022220a0a5b72656d6f7465735d0a2020626f6f7473747261705f75726c203d2022687474703a2f2f6c6f63616c686f73743a39303133220a202063657274735f75726c203d2022687474703a2f2f6c6f63616c686f73743a39303139220a2020687474705f616461707465725f75726c203d2022687474703a2f2f6c6f63616c686f73742f68747470220a20206d73675f636f6e74656e745f74797065203d20226170706c69636174696f6e2f6a736f6e220a20207265616465725f75726c203d2022687474703a2f2f6c6f63616c686f7374220a20207468696e67735f75726c203d2022687474703a2f2f6c6f63616c686f73743a39303030220a2020746c735f766572696669636174696f6e203d2066616c73650a202075736572735f75726c203d2022687474703a2f2f6c6f63616c686f73743a39303032220a"

func init() {
	if os.Getenv("GOBIN") != "" {
		defaultConfigPath = os.Getenv("GOBIN") + "/config.toml"
	}
}

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

	config, err := read(ConfigPath)
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
			return sdkConf, errors.Wrap(errUintConv, err)
		}
		Offset = offset
	}

	if config.Filter.Limit != "" {
		limit, err := strconv.ParseUint(config.Filter.Limit, 10, 64)
		if err != nil {
			return sdkConf, errors.Wrap(errUintConv, err)
		}
		Limit = limit
	}

	if config.Filter.Name != "" {
		Name = config.Filter.Name
	}

	if config.Filter.RawOutput != "" {
		rawOutput, err := strconv.ParseBool(config.Filter.RawOutput)
		if err != nil {
			return sdkConf, errors.Wrap(errBoolConv, err)
		}

		RawOutput = rawOutput
	}
	return nil
}

// New config command to store params to local TOML file.
func NewConfigCmd() *cobra.Command {
	return &cobra.Command{
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

			if err := setConfigValue(key, value); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	}
}

func setConfigValue(key string, value string) error {
	config, err := read(ConfigPath)
	if err != nil {
		return errors.Wrap(errUseExistConf, err)
	}

	if strings.Contains(key, "url") {
		u, err := url.Parse(value)
		if err != nil {
			return errors.Wrap(errURLParseFail, err)
		}
		if u.Scheme == "" || u.Host == "" {
			return err
		}
		if strings.HasPrefix(u.Scheme, "http") || strings.HasPrefix(u.Scheme, "https") {
			return errors.Wrap(errInvalidURL, err)
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
		return errNoKey
	}

	fieldValue := reflect.ValueOf(fieldPtr).Elem()

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return errors.Wrap(errInvalidInt, err)
		}
		fieldValue.SetInt(int64(intValue))
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Wrap(errInvalidBool, err)
		}
		fieldValue.SetBool(boolValue)
	default:
		return errors.Wrap(errUnsupportedKeyValue, err)
	}

	buf, err := toml.Marshal(config)
	if err != nil {
		return errors.Wrap(errMarshal, err)
	}

	err = os.WriteFile(ConfigPath, buf, 0644)
	if err != nil {
		return errors.Wrap(errWritingConfig, err)
	}

	return nil
}
