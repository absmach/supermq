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

	if err := errors.CheckError(resp, http.StatusCreated); err != nil {
		return "", err
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

	if err := errors.CheckError(resp, http.StatusOK); err != nil {
		return User{}, err
	}

	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return User{}, errors.NewSDKError(err.Error())
	}

	return u, nil
}

func (sdk mfSDK) Users(token string, pm PageMetadata) (UsersPage, errors.SDKError) {
	var url string
	var err error

	if url, err = sdk.withQueryParams(sdk.usersURL, usersEndpoint, pm); err != nil {
		return UsersPage{}, errors.NewSDKError(err.Error())
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

	if err := errors.CheckError(resp, http.StatusOK); err != nil {
		return UsersPage{}, err
	}

	var up UsersPage
	if err := json.NewDecoder(resp.Body).Decode(&up); err != nil {
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

	if err := errors.CheckError(resp, http.StatusCreated); err != nil {
		return "", err
	}

	var tr tokenRes
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
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
