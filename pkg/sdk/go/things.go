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
		return "", errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, thingsEndpoint)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", errors.NewSDKError(err)
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return "", errors.NewSDKError(err)
	}
	defer resp.Body.Close()

	if err := errors.CheckError(resp, http.StatusCreated); err != nil {
		return "", err
	}

	id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", thingsEndpoint))
	return id, nil
}

func (sdk mfSDK) CreateThings(things []Thing, token string) ([]Thing, errors.SDKError) {
	data, err := json.Marshal(things)
	if err != nil {
		return []Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, "bulk")

	body, sdkerr := sdk.sendRequestAndGetBodyOrError(http.MethodPost, url, data, token, string(CTJSON), http.StatusCreated)
	if sdkerr != nil {
		return []Thing{}, sdkerr
	}

	var ctr createThingsRes
	if err := json.Unmarshal(body, &ctr); err != nil {
		return []Thing{}, errors.NewSDKError(err)
	}

	return ctr.Things, nil
}

func (sdk mfSDK) Things(token string, pm PageMetadata) (ThingsPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.thingsURL, thingsEndpoint, pm)

	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err)
	}

	body, sdkerr := sdk.sendRequestAndGetBodyOrError(http.MethodGet, url, nil, token, string(CTJSON), http.StatusOK)
	if sdkerr != nil {
		return ThingsPage{}, sdkerr
	}

	var tp ThingsPage
	if err := json.Unmarshal(body, &tp); err != nil {
		return ThingsPage{}, errors.NewSDKError(err)
	}

	return tp, nil
}

func (sdk mfSDK) ThingsByChannel(token, chanID string, offset, limit uint64, disconn bool) (ThingsPage, errors.SDKError) {
	url := fmt.Sprintf("%s/channels/%s/things?offset=%d&limit=%d&disconnected=%t", sdk.thingsURL, chanID, offset, limit, disconn)

	body, err := sdk.sendRequestAndGetBodyOrError(http.MethodGet, url, nil, token, string(CTJSON), http.StatusOK)
	if err != nil {
		return ThingsPage{}, err
	}

	var tp ThingsPage
	if err := json.Unmarshal(body, &tp); err != nil {
		return ThingsPage{}, errors.NewSDKError(err)
	}

	return tp, nil
}

func (sdk mfSDK) Thing(id, token string) (Thing, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, id)

	body, err := sdk.sendRequestAndGetBodyOrError(http.MethodGet, url, nil, token, string(CTJSON), http.StatusOK)
	if err != nil {
		return Thing{}, err
	}

	var t Thing
	if err := json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}

func (sdk mfSDK) UpdateThing(t Thing, token string) errors.SDKError {
	data, err := json.Marshal(t)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, t.ID)

	_, sdkerr := sdk.sendRequestAndGetBodyOrError(http.MethodPut, url, data, token, string(CTJSON), http.StatusOK)
	return sdkerr
}

func (sdk mfSDK) DeleteThing(id, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, id)

	_, err := sdk.sendRequestAndGetBodyOrError(http.MethodDelete, url, nil, token, string(CTJSON), http.StatusNoContent)
	return err
}

func (sdk mfSDK) IdentifyThing(key string) (string, errors.SDKError) {
	idReq := identifyThingReq{Token: key}
	data, err := json.Marshal(idReq)
	if err != nil {
		return "", errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, identifyEndpoint)

	body, sdkerr := sdk.sendRequestAndGetBodyOrError(http.MethodPost, url, data, "", string(CTJSON), http.StatusOK)
	if sdkerr != nil {
		return "", sdkerr
	}

	var i identifyThingResp
	if err := json.Unmarshal(body, &i); err != nil {
		return "", errors.NewSDKError(err)
	}

	return i.ID, nil
}

func (sdk mfSDK) Connect(connIDs ConnectionIDs, token string) errors.SDKError {
	data, err := json.Marshal(connIDs)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, connectEndpoint)

	_, sdkerr := sdk.sendRequestAndGetBodyOrError(http.MethodPost, url, data, token, string(CTJSON), http.StatusOK)
	return sdkerr
}

func (sdk mfSDK) DisconnectThing(thingID, chanID, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.thingsURL, channelsEndpoint, chanID, thingsEndpoint, thingID)

	_, err := sdk.sendRequestAndGetBodyOrError(http.MethodDelete, url, nil, token, string(CTJSON), http.StatusNoContent)
	return err
}
