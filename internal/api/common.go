// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/bootstrap"
	"github.com/absmach/magistrala/internal/apiutil"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/gofrs/uuid"
)

const (
	MemberKindKey    = "member_kind"
	PermissionKey    = "permission"
	RelationKey      = "relation"
	StatusKey        = "status"
	OffsetKey        = "offset"
	OrderKey         = "order"
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
	ListPerms        = "list_perms"
	VisibilityKey    = "visibility"
	SharedByKey      = "shared_by"
	TokenKey         = "token"
	DefPermission    = "view"
	DefTotal         = uint64(100)
	DefOffset        = 0
	DefOrder         = "updated_at"
	DefDir           = "asc"
	DefLimit         = 10
	DefLevel         = 0
	DefStatus        = "enabled"
	DefClientStatus  = mgclients.Enabled
	DefGroupStatus   = mgclients.Enabled
	DefListPerms     = false
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
	if ar, ok := response.(magistrala.Response); ok {
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
	var wrapper error
	if errors.Contains(err, apiutil.ErrValidation) {
		wrapper, err = errors.Unwrap(err)
	}

	w.Header().Set("Content-Type", ContentType)
	switch err.(type) {
	case *svcerr.AuthorizationError, *svcerr.DomainAuthorizationError, *bootstrap.ExternalKeyError, *bootstrap.ExternalKeySecureError:
		err = unwrap(err)
		w.WriteHeader(http.StatusForbidden)

	case *svcerr.AuthenticationError, *apiutil.BearerTokenError, *svcerr.LoginError:
		err = unwrap(err)
		w.WriteHeader(http.StatusUnauthorized)
	case *svcerr.MalformedEntityError,
		*apiutil.MalformedPolicyError,
		*apiutil.MissingSecretError,
		*errors.MalformedEntityError,
		*apiutil.MissingIDError,
		*apiutil.MissingNameError,
		*apiutil.MissingEmailError,
		*apiutil.MissingHostError,
		*apiutil.InvalidResetPassError,
		*apiutil.EmptyListError,
		*apiutil.MissingMemberKindError,
		*apiutil.MissingMemberTypeError,
		*apiutil.LimitSizeError,
		*apiutil.BearerKeyError,
		*svcerr.InvalidStatusError,
		*apiutil.NameSizeError,
		*apiutil.InvalidIDFormatError,
		*apiutil.InvalidQueryParamsError,
		*apiutil.MissingRelationError,
		*apiutil.ValidationError,
		*apiutil.MissingIdentityError,
		*apiutil.MissingPassError,
		*apiutil.MissingConfPassError,
		*apiutil.PasswordFormatError,
		*svcerr.InvalidRoleError,
		*svcerr.InvalidPolicyError,
		*apiutil.InvitationStateError,
		*apiutil.InvalidAPIKeyError,
		*svcerr.ViewEntityError,
		*apiutil.BootstrapStateError,
		*apiutil.MissingCertDataError,
		*apiutil.InvalidContactError,
		*apiutil.InvalidTopicError,
		*bootstrap.AddBootstrapError,
		*apiutil.InvalidCertDataError,
		*apiutil.EmptyMessageError:
		err = unwrap(err)
		w.WriteHeader(http.StatusBadRequest)

	case *svcerr.CreateEntityError,
		*svcerr.UpdateEntityError,
		*svcerr.RemoveEntityError,
		*svcerr.EnableClientError:
		err = unwrap(err)
		w.WriteHeader(http.StatusUnprocessableEntity)

	case *svcerr.NotFoundError,
		*bootstrap.BootstrapError:
		err = unwrap(err)
		w.WriteHeader(http.StatusNotFound)

	case *errors.StatusAlreadyAssignedError,
		*svcerr.ConflictError:
		err = unwrap(err)
		w.WriteHeader(http.StatusConflict)

	case *apiutil.UnsupportedContentTypeError:
		err = unwrap(err)
		w.WriteHeader(http.StatusUnsupportedMediaType)

	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if wrapper != nil {
		err = errors.Wrap(wrapper, err)
	}

	if errorVal, ok := err.(errors.Error); ok {
		if err := json.NewEncoder(w).Encode(errorVal); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func unwrap(err error) error {
	wrapper, err := errors.Unwrap(err)
	if wrapper != nil {
		return wrapper
	}
	return err
}
