// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"io"
	"net/url"
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
}

type config struct {
	Remotes   remotes `toml:"remotes"`
	Filter    filter  `toml:"filter"`
	Channel   channel `toml:"channel"`
	UserToken string  `toml:"user_token"`
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
	fileErr                = errors.New("file error")
	defaultConfigPath      = "./config.toml"
)

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

// Get config parameters from the config file.
func ParseConfig(sdkConf mfxsdk.Config) (mfxsdk.Config, error) {
	if ConfigPath == "" {
		ConfigPath = defaultConfigPath
	}

	_, err := os.Stat(ConfigPath)

	// If the config file does n3ot exist, create it.
	switch {
	case err == os.ErrNotExist:
		// Create the config file with default values
		defaultConfig := config{
			Channel: channel{},
			Filter: filter{
				Offset:    "",
				Limit:     "",
				RawOutput: "",
				Name:      "",
				Contact:   "",
				Email:     "",
				Metadata:  "",
			},
			Remotes: remotes{
				ThingsURL:       "http://localhost:9000",
				UsersURL:        "http://localhost:9002",
				ReaderURL:       "http://localhost",
				HTTPAdapterURL:  "http://localhost/http:9016",
				BootstrapURL:    "http://localhost",
				CertsURL:        "https://localhost:9019",
				TLSVerification: false,
			},
		}
		buf, err := toml.Marshal(defaultConfig)
		if err != nil {
			return sdkConf, errors.Wrap(errMarshal, err)
		}
		err = os.WriteFile(ConfigPath, buf, 0644)
		if err != nil {
			return sdkConf, errors.Wrap(errWritingConfig, err)
		}
	case err != nil:
		return sdkConf, errors.Wrap(fileErr, err)
	}

	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		return sdkConf, errors.Wrap(errConfigNotFound, err)
	}

	config, err := read(ConfigPath)
	if err != nil {
		return sdkConf, errors.Wrap(errReadFail, err)
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

	sdkConf.ThingsURL = config.Remotes.ThingsURL
	sdkConf.UsersURL = config.Remotes.UsersURL
	sdkConf.ReaderURL = config.Remotes.ReaderURL
	sdkConf.HTTPAdapterURL = config.Remotes.HTTPAdapterURL
	sdkConf.BootstrapURL = config.Remotes.BootstrapURL
	sdkConf.CertsURL = config.Remotes.CertsURL

	return sdkConf, nil
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

			if err := setConfigValue(args[0], args[1]); err != nil {
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
			return errors.Wrap(errInvalidURL, err)
		}
		if u.Scheme == "" || u.Host == "" {
			return errInvalidURL
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.Wrap(errURLParseFail, err)
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
		"metadata":         &config.Filter.Metadata,
		"tls_verification": &config.Remotes.TLSVerification,
		"user_token":       &config.UserToken,
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
