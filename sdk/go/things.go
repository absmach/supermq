//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/mainflux/mainflux/things"
)

const thingsEndpoint = "things"

// CreateThing - creates new thing and generates thing UUID
func CreateThing(data, token string) (string, error) {
	url := fmt.Sprintf("%s/%s", serverAddr, thingsEndpoint)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		return "", err
	}

	resp, err := sendRequest(req, token, contentTypeJSON)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("%d", resp.StatusCode)
	}

	return resp.Header.Get("Location"), nil
}

// GetThings - gets all things
func GetThings(token string) ([]things.Thing, error) {
	url := fmt.Sprintf("%s/%s?offset=%s&limit=%s",
		serverAddr, thingsEndpoint, strconv.Itoa(offset), strconv.Itoa(limit))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := sendRequest(req, token, contentTypeJSON)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}

	l := listThingsRes{}
	if err := json.Unmarshal(body, &l); err != nil {
		return nil, err
	}

	return l.Things, nil
}

// GetThing - gets thing by ID
func GetThing(id, token string) (things.Thing, error) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEndpoint, id)
	println(url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return things.Thing{}, err
	}

	resp, err := sendRequest(req, token, contentTypeJSON)
	if err != nil {
		return things.Thing{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return things.Thing{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return things.Thing{}, fmt.Errorf("%d", resp.StatusCode)
	}

	t := things.Thing{}
	if err := json.Unmarshal(body, &t); err != nil {
		return things.Thing{}, err
	}

	return t, nil
}

// UpdateThing - updates thing by ID
func UpdateThing(id, data, token string) error {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEndpoint, id)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := sendRequest(req, token, contentTypeJSON)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil
}

// DeleteThing - removes thing
func DeleteThing(id, token string) error {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEndpoint, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := sendRequest(req, token, contentTypeJSON)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil
}

// ConnectThing - connect thing to a channel
func ConnectThing(cliID, chanID, token string) error {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, channelsEndpoint, chanID, thingsEndpoint, cliID)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	resp, err := sendRequest(req, token, contentTypeJSON)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil
}

// DisconnectThing - connect thing to a channel
func DisconnectThing(cliID, chanID, token string) error {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, channelsEndpoint, chanID, thingsEndpoint, cliID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := sendRequest(req, token, contentTypeJSON)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%d", resp.StatusCode)
	}

	return nil
}
