// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/absmach/supermq/users"
)

type normalizedUser struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Picture   string `json:"picture"`
}

func NormalizeUser(data []byte, provider string) (users.User, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return users.User{}, err
	}

	normalized := normalizeProfile(raw)

	userBytes, err := json.Marshal(normalized)
	if err != nil {
		return users.User{}, err
	}

	var user normalizedUser
	if err := json.Unmarshal(userBytes, &user); err != nil {
		return users.User{}, err
	}

	if err := validateUser(user); err != nil {
		return users.User{}, err
	}

	return users.User{
		ID:             user.ID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		ProfilePicture: user.Picture,
		Metadata:       users.Metadata{"oauth_provider": provider},
	}, nil
}

func normalizeProfile(raw map[string]any) map[string]any {
	normalized := make(map[string]any)

	keyMap := map[string][]string{
		"id":         {"id"},
		"first_name": {"given_name", "first_name", "givenName", "firstname"},
		"last_name":  {"family_name", "last_name", "familyName", "lastname"},
		"username":   {"username", "user_name", "userName"},
		"email":      {"email", "email_address", "emailAddress"},
		"picture":    {"picture", "profile_picture", "profilePicture", "avatar"},
	}

	for stdKey, variants := range keyMap {
		for _, variant := range variants {
			if val, ok := raw[variant]; ok {
				normalized[stdKey] = val
				break
			}
		}
	}

	return normalized
}

func validateUser(user normalizedUser) error {
	var missing []string
	if user.ID == "" {
		missing = append(missing, "id")
	}
	if user.FirstName == "" {
		missing = append(missing, "first_name")
	}
	if user.LastName == "" {
		missing = append(missing, "last_name")
	}
	if user.Email == "" {
		missing = append(missing, "email")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// {"time":"2025-10-04T01:49:45.022963204Z","level":"WARN","msg":"Create domain failed","duration":"10.300227ms","request_id":"09f9b3f6-5be5-4d80-b5db-f037bf90e1ac","domain":{"id":"","name":"d1","route":"d1"},"error":"failed to create entity : entity already exists : ERROR: duplicate key value violates unique constraint \"domains_alias_key\" (SQLSTATE 23505)"}
// {"time":"2025-10-04T01:51:00.207623315Z","level":"WARN","msg":"Register user failed","duration":"706.493278ms","request_id":"aec54295-bac9-4006-9710-7922f6492934","user":{"username":"arvindh","first_name":"Arvindh","last_name":"M"},"error":"failed to create entity : entity already exists : ERROR: duplicate key value violates unique constraint \"clients_identity_key\" (SQLSTATE 23505)"}
// {"time":"2025-10-04T01:52:56.130021294Z","level":"WARN","msg":"Register user failed","duration":"609.641056ms","request_id":"bedb4805-10ec-4811-b90a-44345cbd7392","user":{"username":"arvindh","first_name":"Arvindh","last_name":"M"},"error":"failed to create entity : entity already exists : ERROR: duplicate key value violates unique constraint \"clients_username_key\" (SQLSTATE 23505)"}
// 2025/10/04 01:51:08 traces export: Post "http://mg-jaeger-collector:4318/v1/traces": processor export timeout
