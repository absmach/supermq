// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
)

func (sdk mfSDK) Health() (mainflux.HealthInfo, errors.SDKError) {
	url := fmt.Sprintf("%s/health", sdk.thingsURL)

	resp, err := sdk.client.Get(url)
	if err != nil {
		return mainflux.HealthInfo{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return mainflux.HealthInfo{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return mainflux.HealthInfo{}, encodeError(body, resp.StatusCode)
	}

	var h mainflux.HealthInfo
	if err := json.Unmarshal(body, &h); err != nil {
		return mainflux.HealthInfo{}, errors.NewSDKError(err.Error())
	}

	return h, nil
}
