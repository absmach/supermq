// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-zoo/bone"
	internalerr "github.com/mainflux/mainflux/internal/errors"
	"github.com/mainflux/mainflux/pkg/errors"
)

// ReadUintQuery get the uint64 value for a given http query parameter
func ReadUintQuery(r *http.Request, key string, def uint64) (uint64, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return 0, internalerr.ErrInvalidQueryParams
	}

	if len(vals) == 0 {
		return def, nil
	}

	strval := vals[0]
	val, err := strconv.ParseUint(strval, 10, 64)
	if err != nil {
		return 0, internalerr.ErrInvalidQueryParams
	}

	return val, nil
}

// ReadStringQuery get the string value for a given http query parameter
func ReadStringQuery(r *http.Request, key string) (string, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return "", internalerr.ErrInvalidQueryParams
	}

	if len(vals) == 0 {
		return "", nil
	}

	return vals[0], nil
}

// ReadMetadataQuery get the map[string]interface value for a given http query parameter
func ReadMetadataQuery(r *http.Request, key string) (map[string]interface{}, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return nil, internalerr.ErrInvalidQueryParams
	}

	if len(vals) == 0 {
		return nil, nil
	}

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(vals[0]), &m)
	if err != nil {
		return nil, errors.Wrap(internalerr.ErrInvalidQueryParams, err)
	}

	return m, nil
}

// ReadBoolQuery get the boolean value for a given http query parameter
func ReadBoolQuery(r *http.Request, key string) (bool, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) == 0 {
		return true, internalerr.ErrNotInQuery
	}

	if len(vals) > 1 {
		return false, internalerr.ErrInvalidQueryParams
	}

	b, err := strconv.ParseBool(vals[0])
	if err != nil {
		return false, internalerr.ErrInvalidQueryParams
	}

	return b, nil
}

// ReadFloatQuery get the float64 value for a given http query parameter
func ReadFloatQuery(r *http.Request, key string) (float64, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return 0, internalerr.ErrInvalidQueryParams
	}

	if len(vals) == 0 {
		return 0, nil
	}

	fval := vals[0]
	val, err := strconv.ParseFloat(fval, 64)
	if err != nil {
		return 0, internalerr.ErrInvalidQueryParams
	}

	return val, nil
}
