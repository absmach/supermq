// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	thingsEndpoint   = "things"
	connectEndpoint  = "connect"
	identifyEndpoint = "identify"
)

type identifyThingReq struct {
	Token string `json:"token,omitempty"`
}

type identifyThingResp struct {
	ID string `json:"id,omitempty"`
}

func (sdk mfSDK) CreateThing(t Thing, token string) (string, errors.SDKError) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, thingsEndpoint)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	if sdkerr := errors.CheckError(resp, http.StatusCreated); sdkerr != nil {
		return "", sdkerr
	}

	id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", thingsEndpoint))
	return id, nil
}

func (sdk mfSDK) CreateThings(things []Thing, token string) ([]Thing, errors.SDKError) {
	data, err := json.Marshal(things)
	if err != nil {
		return []Thing{}, errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, "bulk")

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return []Thing{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return []Thing{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Thing{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		return []Thing{}, encodeError(body, resp.StatusCode)
	}

	var ctr createThingsRes
	if err := json.Unmarshal(body, &ctr); err != nil {
		return []Thing{}, errors.NewSDKError(err.Error())
	}

	return ctr.Things, nil
}

func (sdk mfSDK) Things(token string, pm PageMetadata) (ThingsPage, errors.SDKError) {
	url, sdkerr := sdk.withQueryParams(sdk.thingsURL, thingsEndpoint, pm)

	if sdkerr != nil {
		return ThingsPage{}, sdkerr
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return ThingsPage{}, encodeError(body, resp.StatusCode)
	}

	var tp ThingsPage
	if err := json.Unmarshal(body, &tp); err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}

	return tp, nil
}

func (sdk mfSDK) ThingsByChannel(token, chanID string, offset, limit uint64, disconn bool) (ThingsPage, errors.SDKError) {
	url := fmt.Sprintf("%s/channels/%s/things?offset=%d&limit=%d&disconnected=%t", sdk.thingsURL, chanID, offset, limit, disconn)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return ThingsPage{}, encodeError(body, resp.StatusCode)
	}

	var tp ThingsPage
	if err := json.Unmarshal(body, &tp); err != nil {
		return ThingsPage{}, errors.NewSDKError(err.Error())
	}

	return tp, nil
}

func (sdk mfSDK) Thing(id, token string) (Thing, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Thing{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return Thing{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Thing{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return Thing{}, encodeError(body, resp.StatusCode)
	}

	var t Thing
	if err := json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err.Error())
	}

	return t, nil
}

func (sdk mfSDK) UpdateThing(t Thing, token string) errors.SDKError {
	data, err := json.Marshal(t)
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, t.ID)

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

func (sdk mfSDK) DeleteThing(id, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, id)

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

func (sdk mfSDK) IdentifyThing(key string) (string, errors.SDKError) {
	idReq := identifyThingReq{Token: key}
	data, err := json.Marshal(idReq)
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, identifyEndpoint)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, "", string(CTJSON))
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return "", encodeError(body, resp.StatusCode)
	}

	var i identifyThingResp
	if err := json.Unmarshal(body, &i); err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	return i.ID, nil
}

func (sdk mfSDK) Connect(connIDs ConnectionIDs, token string) errors.SDKError {
	data, err := json.Marshal(connIDs)
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, connectEndpoint)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
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

func (sdk mfSDK) DisconnectThing(thingID, chanID, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.thingsURL, channelsEndpoint, chanID, thingsEndpoint, thingID)
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
