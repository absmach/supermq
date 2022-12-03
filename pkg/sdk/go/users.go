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
	usersEndpoint    = "users"
	tokensEndpoint   = "tokens"
	passwordEndpoint = "password"
	membersEndpoint  = "members"
)

func (sdk mfSDK) CreateUser(token string, u User) (string, errors.SDKError) {
	data, err := json.Marshal(u)
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, usersEndpoint)

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

	id := strings.TrimPrefix(resp.Header.Get("Location"), fmt.Sprintf("/%s/", usersEndpoint))
	return id, nil
}

func (sdk mfSDK) User(userID, token string) (User, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, userID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return User{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return User{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return User{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return User{}, encodeError(body, resp.StatusCode)
	}

	var u User
	if err := json.Unmarshal(body, &u); err != nil {
		return User{}, errors.NewSDKError(err.Error())
	}

	return u, nil
}

func (sdk mfSDK) Users(token string, pm PageMetadata) (UsersPage, errors.SDKError) {
	url, sdkerr := sdk.withQueryParams(sdk.usersURL, usersEndpoint, pm)

	if sdkerr != nil {
		return UsersPage{}, errors.NewSDKError(sdkerr.Error())
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return UsersPage{}, errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return UsersPage{}, errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return UsersPage{}, errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return UsersPage{}, encodeError(body, resp.StatusCode)
	}
	var up UsersPage
	if err := json.Unmarshal(body, &up); err != nil {
		return UsersPage{}, errors.NewSDKError(err.Error())
	}

	return up, nil
}

func (sdk mfSDK) CreateToken(user User) (string, errors.SDKError) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, tokensEndpoint)

	resp, err := sdk.client.Post(url, string(CTJSON), bytes.NewReader(data))
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		return "", encodeError(body, resp.StatusCode)
	}

	var tr tokenRes
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", errors.NewSDKError(err.Error())
	}

	return tr.Token, nil
}

func (sdk mfSDK) UpdateUser(u User, token string) errors.SDKError {
	data, err := json.Marshal(u)
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, usersEndpoint)

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

func (sdk mfSDK) UpdatePassword(oldPass, newPass, token string) errors.SDKError {
	ur := UserPasswordReq{
		OldPassword: oldPass,
		Password:    newPass,
	}
	data, err := json.Marshal(ur)
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, passwordEndpoint)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
	if err != nil {
		return errors.NewSDKError(err.Error())
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return errors.NewSDKError(err.Error())
	}
	defer resp.Body.Close()

	return errors.CheckError(resp, http.StatusCreated)
}

func (sdk mfSDK) EnableUser(id, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/enable", sdk.usersURL, usersEndpoint, id)

	req, err := http.NewRequest(http.MethodPost, url, nil)
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

func (sdk mfSDK) DisableUser(id, token string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/disable", sdk.usersURL, usersEndpoint, id)

	req, err := http.NewRequest(http.MethodPost, url, nil)
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
