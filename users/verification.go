// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/absmach/supermq/pkg/errors"
)

type verificationToken struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func newVerificationToken(userID, email string) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errors.Wrap(errFailedToEncodeToken, err)
	}

	payload := verificationToken{
		UserID: userID,
		Email:  email,
		Token:  base64.URLEncoding.EncodeToString(randomBytes),
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", errors.Wrap(errFailedToEncodeToken, err)
	}

	return base64.URLEncoding.EncodeToString(jsonBytes), nil
}

func decodeVerificationToken(token string) (verificationToken, error) {
	decodedPayload, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return verificationToken{}, errors.Wrap(errFailedToDecodeToken, err)
	}

	var payload verificationToken
	if err := json.Unmarshal(decodedPayload, &payload); err != nil {
		return verificationToken{}, errInvalidTokenFormat
	}

	return payload, nil
}
