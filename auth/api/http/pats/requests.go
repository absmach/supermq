// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pats

import (
	"encoding/json"
	"strings"
	"time"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/auth"
)

type createPatReq struct {
	token       string
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
}

func (cpr *createPatReq) UnmarshalJSON(data []byte) error {
	var temp struct {
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		Duration    string `json:"duration,omitempty"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	duration, err := time.ParseDuration(temp.Duration)
	if err != nil {
		return err
	}
	cpr.Name = temp.Name
	cpr.Description = temp.Description
	cpr.Duration = duration
	return nil
}

func (req createPatReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if strings.TrimSpace(req.Name) == "" {
		return apiutil.ErrMissingName
	}

	return nil
}

type retrievePatReq struct {
	token string
	id    string
}

func (req retrievePatReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type updatePatNameReq struct {
	token string
	id    string
	Name  string `json:"name,omitempty"`
}

func (req updatePatNameReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if strings.TrimSpace(req.Name) == "" {
		return apiutil.ErrMissingName
	}
	return nil
}

type updatePatDescriptionReq struct {
	token       string
	id          string
	Description string `json:"description,omitempty"`
}

func (req updatePatDescriptionReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if strings.TrimSpace(req.Description) == "" {
		return apiutil.ErrMissingDescription
	}
	return nil
}

type listPatsReq struct {
	token  string
	offset uint64
	limit  uint64
}

func (req listPatsReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	return nil
}

type deletePatReq struct {
	token string
	id    string
}

func (req deletePatReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type resetPatSecretReq struct {
	token    string
	id       string
	Duration time.Duration `json:"duration,omitempty"`
}

func (rspr *resetPatSecretReq) UnmarshalJSON(data []byte) error {
	var temp struct {
		Duration string `json:"duration,omitempty"`
	}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	rspr.Duration, err = time.ParseDuration(temp.Duration)
	if err != nil {
		return err
	}
	return nil
}

func (req resetPatSecretReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type revokePatSecretReq struct {
	token string
	id    string
}

func (req revokePatSecretReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type scopeEntryReq struct {
	token  string
	id     string
	Scopes []auth.Scope `json:"scopes,omitempty"`
}

func (aser *scopeEntryReq) UnmarshalJSON(data []byte) error {
	type Alias scopeEntryReq
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(aser),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	return nil
}

func (req scopeEntryReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingPATID
	}

	if len(req.Scopes) == 0 {
		return apiutil.ErrValidation
	}

	return nil
}

type clearAllScopeEntryReq struct {
	token string
	id    string
}

func (req clearAllScopeEntryReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type listScopesReq struct {
	token  string
	offset uint64
	limit  uint64
	patID  string
}

func (req listScopesReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.patID == "" {
		return apiutil.ErrMissingPATID
	}
	return nil
}
