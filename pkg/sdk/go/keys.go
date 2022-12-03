// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
)

type keyReq struct {
	Type     uint32        `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

const keysEndpoint = "keys"

const (
	// LoginKey is temporary User key received on successfull login.
	LoginKey uint32 = iota
	// RecoveryKey represents a key for resseting password.
	RecoveryKey
	// APIKey enables the one to act on behalf of the user.
	APIKey
)

func (sdk mfSDK) Issue(token string, d time.Duration) (KeyRes, errors.SDKError) {
	datareq := keyReq{Type: APIKey, Duration: d}
	data, err := json.Marshal(datareq)
	if err != nil {
		return KeyRes{}, errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.authURL, keysEndpoint)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return KeyRes{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return KeyRes{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return KeyRes{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		return KeyRes{}, encodeError(body, resp.StatusCode)
	}

	var key KeyRes
	if err := json.Unmarshal(body, &key); err != nil {
		return KeyRes{}, errors.NewSDKError(err.Error())
	}

	return key, nil
}

func (sdk mfSDK) Revoke(id, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s", sdk.authURL, keysEndpoint, id)
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

func (sdk mfSDK) RetrieveKey(id, token string) (retrieveKeyRes, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.authURL, keysEndpoint, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return retrieveKeyRes{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return retrieveKeyRes{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return retrieveKeyRes{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return retrieveKeyRes{}, encodeError(body, resp.StatusCode)
	}

	var key retrieveKeyRes
	if err := json.Unmarshal(body, &key); err != nil {
		return retrieveKeyRes{}, errors.NewSDKError(err.Error())
	}

	return key, nil
}
