package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	usersEndpoint   = "users"
	enableEndpoint  = "enable"
	disableEndpoint = "disable"
	tokensEndpoint  = "tokens/issue"
	membersEndpoint = "members"
)

// MembershipsPage contains page related metadata as well as list of memberships that
// belong to this page.
type MembershipsPage struct {
	PageMetadata
	Memberships []Group
}

// updateClientSecretReq is used to update the client secret
type updateClientSecretReq struct {
	OldSecret string `json:"old_secret,omitempty"`
	NewSecret string `json:"new_secret,omitempty"`
}

// updateClientIdentityReq is used to update the client identity
type updateClientIdentityReq struct {
	token    string
	id       string
	Identity string `json:"identity,omitempty"`
}

// CreateClient creates a new client returning its id.
func (sdk mfSDK) CreateUser(user User, token string) (string, errors.SDKError) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, usersEndpoint)

	headers, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusCreated)
	if sdkerr != nil {
		return "", sdkerr
	}

	id := strings.TrimPrefix(headers.Get("Location"), fmt.Sprintf("/%s/", usersEndpoint))
	return id, nil
}

// Users returns page of users.
func (sdk mfSDK) Users(pm PageMetadata, token string) (UsersPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.usersURL, usersEndpoint, pm)
	if err != nil {
		return UsersPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return UsersPage{}, sdkerr
	}

	var cp UsersPage
	if err := json.Unmarshal(body, &cp); err != nil {
		return UsersPage{}, errors.NewSDKError(err)
	}

	return cp, nil
}

// ListMembers retrieves everything that is assigned to a group identified by groupID.
func (sdk mfSDK) ListMembers(groupID string, meta PageMetadata, token string) (MembersPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.usersURL, fmt.Sprintf("%s/%s/%s", groupsEndpoint, groupID, membersEndpoint), meta)
	if err != nil {
		return MembersPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return MembersPage{}, sdkerr
	}

	var mp MembersPage
	if err := json.Unmarshal(body, &mp); err != nil {
		return MembersPage{}, errors.NewSDKError(err)
	}

	return mp, nil
}

// User returns user object by id.
func (sdk mfSDK) User(id, token string) (User, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, id)

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

// UpdateUser updates existing user.
func (sdk mfSDK) UpdateUser(user User, token string) errors.SDKError {
	data, err := json.Marshal(user)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, user.ID)

	_, _, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)

	return sdkerr
}

// UpdateUserTags updates the user's tags.
func (sdk mfSDK) UpdateUserTags(user User, token string) errors.SDKError {
	data, err := json.Marshal(user)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/tags", sdk.usersURL, usersEndpoint, user.ID)

	_, _, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)

	return sdkerr
}

// UpdateUserIdentity updates the user's identity
func (sdk mfSDK) UpdateUserIdentity(user User, token string) errors.SDKError {
	ucir := updateClientIdentityReq{token: token, id: user.ID, Identity: user.Credentials.Identity}

	data, err := json.Marshal(ucir)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/identity", sdk.usersURL, usersEndpoint, user.ID)

	_, _, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)

	return sdkerr
}

// UpdatePassword updates user password.
func (sdk mfSDK) UpdatePassword(id, oldPass, newPass, token string) errors.SDKError {
	var ucsr = updateClientSecretReq{OldSecret: oldPass, NewSecret: newPass}

	data, err := json.Marshal(ucsr)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/secret", sdk.usersURL, usersEndpoint, id)

	_, _, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)

	return sdkerr
}

// UpdateUserOwner updates the user's owner.
func (sdk mfSDK) UpdateUserOwner(user User, token string) errors.SDKError {
	data, err := json.Marshal(user)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/owner", sdk.usersURL, usersEndpoint, user.ID)

	_, _, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)

	return sdkerr
}

// EnableUser changes the status of the user to enabled.
func (sdk mfSDK) EnableUser(id, token string) errors.SDKError {
	return sdk.changeClientStatus(token, id, enableEndpoint)
}

// DisableUser changes the status of the user to disabled.
func (sdk mfSDK) DisableUser(id, token string) errors.SDKError {
	return sdk.changeClientStatus(token, id, disableEndpoint)
}

func (sdk mfSDK) changeClientStatus(token, id, status string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.usersURL, usersEndpoint, id, status)
	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), nil, http.StatusOK)

	return sdkerr
}

// CreateToken receives credentials and returns user token.
func (sdk mfSDK) CreateToken(user User) (string, errors.SDKError) {
	var treq = tokenReq{
		Identity: user.Credentials.Identity,
		Secret:   user.Credentials.Secret,
	}
	data, err := json.Marshal(treq)
	if err != nil {
		return "", errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, tokensEndpoint)

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, "", string(CTJSON), data, http.StatusCreated)
	if sdkerr != nil {
		return "", sdkerr
	}
	var tr tokenRes
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", errors.NewSDKError(err)
	}

	return tr.AccessToken, nil
}
