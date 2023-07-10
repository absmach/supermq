// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/internal/postgres"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	StatusKey        = "status"
	OffsetKey        = "offset"
	LimitKey         = "limit"
	MetadataKey      = "metadata"
	ParentKey        = "parent_id"
	OwnerKey         = "owner_id"
	ClientKey        = "client"
	IdentityKey      = "identity"
	GroupKey         = "group"
	ActionKey        = "action"
	TagKey           = "tag"
	NameKey          = "name"
	TotalKey         = "total"
	SubjectKey       = "subject"
	ObjectKey        = "object"
	LevelKey         = "level"
	TreeKey          = "tree"
	DirKey           = "dir"
	VisibilityKey    = "visibility"
	SharedByKey      = "shared_by"
	TokenKey         = "token"
	DefTotal         = uint64(100)
	DefOffset        = 0
	DefLimit         = 10
	DefLevel         = 0
	DefStatus        = "enabled"
	DefClientStatus  = mfclients.Enabled
	DefGroupStatus   = mfclients.Enabled
	SharedVisibility = "shared"
	MyVisibility     = "mine"
	AllVisibility    = "all"
	// ContentType represents JSON content type.
	ContentType = "application/json"

	// MaxNameSize limits name size to prevent making them too complex.
	MaxLimitSize = 100
	MaxNameSize  = 1024
	NameOrder    = "name"
	IDOrder      = "id"
	AscDir       = "asc"
	DescDir      = "desc"
)

// ValidateUUID validates UUID format.
func ValidateUUID(extID string) (err error) {
	id, err := uuid.FromString(extID)
	if id.String() != extID || err != nil {
		return apiutil.ErrInvalidIDFormat
	}

	return nil
}

// EncodeResponse encodes successful response.
func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", ContentType)
		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

// EncodeError encodes an error response.
func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", ContentType)
	switch {
	case strings.Contains(err.Error(), apiutil.ErrMalformedEntity.Error()),
		strings.Contains(err.Error(), apiutil.ErrMissingID.Error()),
		strings.Contains(err.Error(), apiutil.ErrEmptyList.Error()),
		strings.Contains(err.Error(), apiutil.ErrMissingMemberType.Error()),
		strings.Contains(err.Error(), apiutil.ErrInvalidSecret.Error()),
		strings.Contains(err.Error(), apiutil.ErrNameSize.Error()):
		w.WriteHeader(http.StatusBadRequest)
	case strings.Contains(err.Error(), errors.ErrAuthentication.Error()):
		w.WriteHeader(http.StatusUnauthorized)
	case strings.Contains(err.Error(), errors.ErrNotFound.Error()):
		w.WriteHeader(http.StatusNotFound)
	case strings.Contains(err.Error(), errors.ErrConflict.Error()):
		w.WriteHeader(http.StatusConflict)
	case strings.Contains(err.Error(), errors.ErrAuthorization.Error()):
		w.WriteHeader(http.StatusForbidden)
	case strings.Contains(err.Error(), postgres.ErrMemberAlreadyAssigned.Error()):
		w.WriteHeader(http.StatusConflict)
	case strings.Contains(err.Error(), apiutil.ErrUnsupportedContentType.Error()):
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case strings.Contains(err.Error(), errors.ErrCreateEntity.Error()),
		strings.Contains(err.Error(), errors.ErrUpdateEntity.Error()),
		strings.Contains(err.Error(), errors.ErrViewEntity.Error()),
		strings.Contains(err.Error(), errors.ErrRemoveEntity.Error()):
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if errorVal, ok := err.(errors.Error); ok {

		errMsg := errorVal.Msg()
		if errorVal.Err() != nil {
			// errMsg = errorVal.Err().Msg()
			errMsg = fmt.Sprintf("%s : %s", errorVal.Msg(), errorVal.Err().Msg())
		}
		if err := json.NewEncoder(w).Encode(apiutil.ErrorRes{Err: errMsg}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
