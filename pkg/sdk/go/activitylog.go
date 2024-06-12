// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
)

const activitiesEndpoint = "activities"

type Activity struct {
	ID         string    `json:"id,omitempty"`
	Operation  string    `json:"operation,omitempty"`
	OccurredAt time.Time `json:"occurred_at,omitempty"`
	Payload    Metadata  `json:"payload,omitempty"`
}

type ActivitiesPage struct {
	Total      uint64     `json:"total"`
	Offset     uint64     `json:"offset"`
	Limit      uint64     `json:"limit"`
	Activities []Activity `json:"activities"`
}

func (sdk mgSDK) Activities(entityID, entityType string, pm PageMetadata, token string) (activities ActivitiesPage, err error) {
	if entityID == "" {
		return ActivitiesPage{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	if entityType == "" {
		return ActivitiesPage{}, errors.NewSDKError(apiutil.ErrMissingEntityType)
	}

	url, err := sdk.withQueryParams(sdk.activitiesURL, fmt.Sprintf("%s/%s/%s", activitiesEndpoint, entityType, entityID), pm)
	if err != nil {
		return ActivitiesPage{}, errors.NewSDKError(err)
	}
	fmt.Println(url)

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return ActivitiesPage{}, sdkerr
	}

	var activitiesPage ActivitiesPage
	if err := json.Unmarshal(body, &activitiesPage); err != nil {
		return ActivitiesPage{}, errors.NewSDKError(err)
	}

	return activitiesPage, nil
}
