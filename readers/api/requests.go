// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"time"

	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/absmach/magistrala/readers"
)

const maxLimitSize = 1000

type listMessagesReq struct {
	chanID   string
	token    string
	key      string
	pageMeta readers.PageMetadata
}

func (req listMessagesReq) validate() error {
	if req.token == "" && req.key == "" {
		return apiutil.ErrBearerToken
	}

	if req.chanID == "" {
		return apiutil.ErrMissingID
	}

	if req.pageMeta.Limit < 1 || req.pageMeta.Limit > maxLimitSize {
		return apiutil.ErrLimitSize
	}

	if req.pageMeta.Comparator != "" &&
		req.pageMeta.Comparator != readers.EqualKey &&
		req.pageMeta.Comparator != readers.LowerThanKey &&
		req.pageMeta.Comparator != readers.LowerThanEqualKey &&
		req.pageMeta.Comparator != readers.GreaterThanKey &&
		req.pageMeta.Comparator != readers.GreaterThanEqualKey {
		return apiutil.ErrInvalidComparator
	}

	validAggregations := map[string]bool{
		"MAX": true,
		"MIN": true,
		"AVG": true,
		"SUM": true,
		"COUNT": true,
		"max": true,
		"min": true,
		"avg": true,
		"sum": true,
		"count": true,
	}

	aggregationValid := validAggregations[req.pageMeta.Aggregation]

	_, err := time.ParseDuration(req.pageMeta.Interval)
	validInterval := err == nil

	if (req.pageMeta.Aggregation != "" || req.pageMeta.Interval != "" || req.pageMeta.To != 0 || req.pageMeta.From != 0) &&
		!aggregationValid {
		return apiutil.ErrInvalidAggregation
	}

	if (req.pageMeta.Aggregation != "" || req.pageMeta.Interval != "" || req.pageMeta.To != 0 || req.pageMeta.From != 0) &&
		!validInterval {
		return apiutil.ErrInvalidInterval
	}

	if (req.pageMeta.Aggregation != "" || req.pageMeta.Interval != "" || req.pageMeta.From != 0) &&
		req.pageMeta.To == 0 {
		return apiutil.ErrMissingTo
	}

	if (req.pageMeta.Aggregation != "" || req.pageMeta.Interval != "" || req.pageMeta.To != 0) &&
		req.pageMeta.From == 0 {
		return apiutil.ErrMissingFrom
	}

	if (req.pageMeta.Aggregation != "" || req.pageMeta.To != 0 || req.pageMeta.From != 0) &&
		req.pageMeta.Interval == "" {
		return apiutil.ErrMissingInterval
	}

	if (req.pageMeta.Interval != "" || req.pageMeta.To != 0 || req.pageMeta.From != 0) &&
		req.pageMeta.Aggregation == "" {
		return apiutil.ErrMissingAggregation
	}

	return nil
}
