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
	Scope       []auth.Scope  `json:"scope,omitempty"`
}

func (cpr *createPatReq) UnmarshalJSON(data []byte) error {
	var temp struct {
		Name        string       `json:"name,omitempty"`
		Description string       `json:"description,omitempty"`
		Duration    string       `json:"duration,omitempty"`
		Scope       []auth.Scope `json:"scope,omitempty"`
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
	cpr.Scope = temp.Scope
	return nil
}

func (req createPatReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if strings.TrimSpace(req.Name) == "" {
		return apiutil.ErrMissingName
	}

	for _, s := range req.Scope {
		if s.EntityId == "" {
			return apiutil.ErrMissingID
		}
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

type addPatScopeEntryReq struct {
	token            string
	id               string
	EntityType       auth.EntityType `json:"entity_type,omitempty"`
	OptionalDomainID string          `json:"optional_domain_id,omitempty"`
	Operation        auth.Operation  `json:"operation,omitempty"`
	EntityIDs        []string        `json:"entity_ids,omitempty"`
}

func (apser *addPatScopeEntryReq) UnmarshalJSON(data []byte) error {
	var temp struct {
		EntityType       string   `json:"entity_type,omitempty"`
		OptionalDomainID string   `json:"optional_domain_id,omitempty"`
		Operation        string   `json:"operation,omitempty"`
		EntityIDs        []string `json:"entity_ids,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	pet, err := auth.ParseEntityType(temp.EntityType)
	if err != nil {
		return err
	}
	op, err := auth.ParseOperation(temp.Operation)
	if err != nil {
		return err
	}
	apser.EntityType = pet
	apser.OptionalDomainID = temp.OptionalDomainID
	apser.Operation = op
	apser.EntityIDs = temp.EntityIDs
	return nil
}

func (req addPatScopeEntryReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type removePatScopeEntryReq struct {
	token            string
	id               string
	EntityType       auth.EntityType `json:"platform_entity_type,omitempty"`
	OptionalDomainID string          `json:"optional_domain_id,omitempty"`
	Operation        auth.Operation  `json:"operation,omitempty"`
	EntityIDs        []string        `json:"entity_ids,omitempty"`
}

func (rpser *removePatScopeEntryReq) UnmarshalJSON(data []byte) error {
	var temp struct {
		EntityType       string   `json:"entity_type,omitempty"`
		OptionalDomainID string   `json:"optional_domain_id,omitempty"`
		Operation        string   `json:"operation,omitempty"`
		EntityIDs        []string `json:"entity_ids,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	pet, err := auth.ParseEntityType(temp.EntityType)
	if err != nil {
		return err
	}
	op, err := auth.ParseOperation(temp.Operation)
	if err != nil {
		return err
	}
	rpser.EntityType = pet
	rpser.OptionalDomainID = temp.OptionalDomainID
	rpser.Operation = op
	rpser.EntityIDs = temp.EntityIDs
	return nil
}

func (req removePatScopeEntryReq) validate() (err error) {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
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
	userID string
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
