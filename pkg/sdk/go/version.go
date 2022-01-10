// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mainflux/mainflux/pkg/errors"
)

type health struct {
	Version string `json:"version"`
}

func (sdk mfSDK) Version() (string, error) {
	url := fmt.Sprintf("%s/health", sdk.thingsURL)

	resp, err := sdk.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Wrap(ErrFetchVersion, errors.New(resp.Status))
	}

	var h health
	if err := json.Unmarshal(body, &h); err != nil {
		return "", err
	}

	return h.Version, nil
}
