// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mainflux/mainflux/pkg/errors"
)

const configsEndpoint = "configs"
const bootstrapEndpoint = "bootstrap"
const whitelistEndpoint = "state"
const bootstrapCertsEndpoint = "configs/certs"

// BootstrapConfig represents Configuration entity. It wraps information about external entity
// as well as info about corresponding Mainflux entities.
// MFThing represents corresponding Mainflux Thing ID.
// MFKey is key of corresponding Mainflux Thing.
// MFChannels is a list of Mainflux Channels corresponding Mainflux Thing connects to.
type BootstrapConfig struct {
	ThingID     string    `json:"thing_id,omitempty"`
	Channels    []string  `json:"channels,omitempty"`
	ExternalID  string    `json:"external_id,omitempty"`
	ExternalKey string    `json:"external_key,omitempty"`
	MFThing     string    `json:"mainflux_id,omitempty"`
	MFChannels  []Channel `json:"mainflux_channels,omitempty"`
	MFKey       string    `json:"mainflux_key,omitempty"`
	Name        string    `json:"name,omitempty"`
	ClientCert  string    `json:"client_cert,omitempty"`
	ClientKey   string    `json:"client_key,omitempty"`
	CACert      string    `json:"ca_cert,omitempty"`
	Content     string    `json:"content,omitempty"`
	State       int       `json:"state,omitempty"`
}

type ConfigUpdateCertReq struct {
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
	CACert     string `json:"ca_cert"`
}

func (sdk mfSDK) AddBootstrap(token string, cfg BootstrapConfig) (string, errors.SDKError) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.bootstrapURL, configsEndpoint)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	if err := errors.CheckError(resp, http.StatusOK, http.StatusCreated); err != nil {
		return "", err
	}

	id := strings.TrimPrefix(resp.Header.Get("Location"), "/things/configs/")
	return id, nil
}

func (sdk mfSDK) Whitelist(token string, cfg BootstrapConfig) errors.SDKError {
	data, err := json.Marshal(BootstrapConfig{State: cfg.State})
	if err != nil {
		return errors.NewSDKError(errors.Wrap(ErrFailedWhitelist, err).Error())
	}

	if cfg.MFThing == "" {
		return ErrFailedWhitelist
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.bootstrapURL, whitelistEndpoint, cfg.MFThing)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return errors.NewSDKError(errors.Wrap(ErrFailedWhitelist, err).Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return errors.NewSDKError(errors.Wrap(ErrFailedWhitelist, err).Error())
	}
	defer resp.Body.Close()

	return errors.CheckError(resp, http.StatusCreated, http.StatusOK)
}

func (sdk mfSDK) ViewBootstrap(token, id string) (BootstrapConfig, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.bootstrapURL, configsEndpoint, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return BootstrapConfig{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return BootstrapConfig{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	if err := errors.CheckError(resp, http.StatusOK); err != nil {
		return BootstrapConfig{}, err
	}

	var bc BootstrapConfig
	if err := json.NewDecoder(resp.Body).Decode(&bc); err != nil {
		return BootstrapConfig{}, errors.NewSDKError(err.Error())
	}

	return bc, nil
}

func (sdk mfSDK) UpdateBootstrap(token string, cfg BootstrapConfig) errors.SDKError {
	data, err := json.Marshal(cfg)
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.bootstrapURL, configsEndpoint, cfg.MFThing)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	return errors.CheckError(resp, http.StatusOK)
}
func (sdk mfSDK) UpdateBootstrapCerts(token, id, clientCert, clientKey, ca string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s", sdk.bootstrapURL, bootstrapCertsEndpoint, id)
	request := ConfigUpdateCertReq{
		ClientCert: clientCert,
		ClientKey:  clientKey,
		CACert:     ca,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return errors.NewSDKError(errors.Wrap(ErrFailedCertUpdate, err).Error())
	}
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	return errors.CheckError(resp, http.StatusOK)
}

func (sdk mfSDK) RemoveBootstrap(token, id string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s", sdk.bootstrapURL, configsEndpoint, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	return errors.CheckError(resp, http.StatusNoContent)
}

func (sdk mfSDK) Bootstrap(externalKey, externalID string) (BootstrapConfig, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.bootstrapURL, bootstrapEndpoint, externalID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return BootstrapConfig{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, externalKey, string(CTJSON))
	if err != nil {
		return BootstrapConfig{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	if err := errors.CheckError(resp, http.StatusOK); err != nil {
		return BootstrapConfig{}, err
	}

	var bc BootstrapConfig
	if err := json.NewDecoder(resp.Body).Decode(&bc); err != nil {
		return BootstrapConfig{}, errors.NewSDKError(err.Error())
	}

	return bc, nil
}
