// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/pkg/errors"
)

const (
	usersEndpoint         = "users"
	enableEndpoint        = "enable"
	disableEndpoint       = "disable"
	issueTokenEndpoint    = "tokens/issue"
	refreshTokenEndpoint  = "tokens/refresh"
	membersEndpoint       = "members"
	PasswordResetEndpoint = "password"
)

// User represents supermq user its credentials.
type User struct {
	ID             string      `json:"id"`
	FirstName      string      `json:"first_name,omitempty"`
	LastName       string      `json:"last_name,omitempty"`
	Email          string      `json:"email,omitempty"`
	Credentials    Credentials `json:"credentials"`
	Tags           []string    `json:"tags,omitempty"`
	Metadata       Metadata    `json:"metadata,omitempty"`
	CreatedAt      time.Time   `json:"created_at,omitempty"`
	UpdatedAt      time.Time   `json:"updated_at,omitempty"`
	Status         string      `json:"status,omitempty"`
	Role           string      `json:"role,omitempty"`
	ProfilePicture string      `json:"profile_picture,omitempty"`
}

func (sdk mgSDK) CreateUser(ctx context.Context, user User, token string) (User, errors.SDKError) {
	data, err := json.Marshal(user)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.usersURL, usersEndpoint)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPost, url, token, data, nil, http.StatusCreated)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user = User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) Users(ctx context.Context, pm PageMetadata, token string) (UsersPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.usersURL, usersEndpoint, pm)
	if err != nil {
		return UsersPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return UsersPage{}, sdkerr
	}

	var cp UsersPage
	if err := json.Unmarshal(body, &cp); err != nil {
		return UsersPage{}, errors.NewSDKError(err)
	}

	return cp, nil
}

func (sdk mgSDK) User(ctx context.Context, id, token string) (User, errors.SDKError) {
	if id == "" {
		return User{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, id)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) UserProfile(ctx context.Context, token string) (User, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/profile", sdk.usersURL, usersEndpoint)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) UpdateUser(ctx context.Context, user User, token string) (User, errors.SDKError) {
	if user.ID == "" {
		return User{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user = User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) UpdateUserTags(ctx context.Context, user User, token string) (User, errors.SDKError) {
	data, err := json.Marshal(user)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/tags", sdk.usersURL, usersEndpoint, user.ID)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user = User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) UpdateUserEmail(ctx context.Context, user User, token string) (User, errors.SDKError) {
	ucir := updateUserEmailReq{token: token, id: user.ID, Email: user.Email}

	data, err := json.Marshal(ucir)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/email", sdk.usersURL, usersEndpoint, user.ID)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user = User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) ResetPasswordRequest(ctx context.Context, email string) errors.SDKError {
	rpr := resetPasswordRequestreq{Email: email}

	data, err := json.Marshal(rpr)
	if err != nil {
		return errors.NewSDKError(err)
	}
	url := fmt.Sprintf("%s/%s/reset-request", sdk.usersURL, PasswordResetEndpoint)

	header := make(map[string]string)
	header["Referer"] = sdk.HostURL

	_, _, sdkerr := sdk.processRequest(ctx, http.MethodPost, url, "", data, header, http.StatusCreated)

	return sdkerr
}

func (sdk mgSDK) ResetPassword(ctx context.Context, password, confPass, token string) errors.SDKError {
	rpr := resetPasswordReq{Token: token, Password: password, ConfPass: confPass}

	data, err := json.Marshal(rpr)
	if err != nil {
		return errors.NewSDKError(err)
	}
	url := fmt.Sprintf("%s/%s/reset", sdk.usersURL, PasswordResetEndpoint)

	_, _, sdkerr := sdk.processRequest(ctx, http.MethodPut, url, token, data, nil, http.StatusCreated)

	return sdkerr
}

func (sdk mgSDK) UpdatePassword(ctx context.Context, oldPass, newPass, token string) (User, errors.SDKError) {
	ucsr := updateUserSecretReq{OldSecret: oldPass, NewSecret: newPass}

	data, err := json.Marshal(ucsr)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/secret", sdk.usersURL, usersEndpoint)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	var user User
	if err = json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) UpdateUserRole(ctx context.Context, user User, token string) (User, errors.SDKError) {
	data, err := json.Marshal(user)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/role", sdk.usersURL, usersEndpoint, user.ID)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user = User{}
	if err = json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) UpdateUsername(ctx context.Context, user User, token string) (User, errors.SDKError) {
	uur := UpdateUsernameReq{id: user.ID, Username: user.Credentials.Username}
	data, err := json.Marshal(uur)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/username", sdk.usersURL, usersEndpoint, user.ID)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user = User{}
	if err = json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) UpdateProfilePicture(ctx context.Context, user User, token string) (User, errors.SDKError) {
	data, err := json.Marshal(user)
	if err != nil {
		return User{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/picture", sdk.usersURL, usersEndpoint, user.ID)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user = User{}
	if err = json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) SearchUsers(ctx context.Context, pm PageMetadata, token string) (UsersPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.usersURL, fmt.Sprintf("%s/search", usersEndpoint), pm)
	if err != nil {
		return UsersPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return UsersPage{}, sdkerr
	}

	var cp UsersPage
	if err := json.Unmarshal(body, &cp); err != nil {
		return UsersPage{}, errors.NewSDKError(err)
	}

	return cp, nil
}

func (sdk mgSDK) EnableUser(ctx context.Context, id, token string) (User, errors.SDKError) {
	return sdk.changeUserStatus(ctx, token, id, enableEndpoint)
}

func (sdk mgSDK) DisableUser(ctx context.Context, id, token string) (User, errors.SDKError) {
	return sdk.changeUserStatus(ctx, token, id, disableEndpoint)
}

func (sdk mgSDK) changeUserStatus(ctx context.Context, token, id, status string) (User, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.usersURL, usersEndpoint, id, status)

	_, body, sdkerr := sdk.processRequest(ctx, http.MethodPost, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return User{}, sdkerr
	}

	user := User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, errors.NewSDKError(err)
	}

	return user, nil
}

func (sdk mgSDK) DeleteUser(ctx context.Context, id, token string) errors.SDKError {
	if id == "" {
		return errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s", sdk.usersURL, usersEndpoint, id)
	_, _, sdkerr := sdk.processRequest(ctx, http.MethodDelete, url, token, nil, nil, http.StatusNoContent)
	return sdkerr
}
