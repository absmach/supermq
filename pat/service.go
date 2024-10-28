// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pat

import (
	"context"
	"encoding/base64"
	"math/rand"
	"strings"
	"time"

	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
)

const (
	randStr            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&&*|+-="
	patPrefix          = "pat"
	patSecretSeparator = "_"
)

var (
	errMalformedPAT        = errors.New("malformed personal access token")
	errFailedToParseUUID   = errors.New("failed to parse string to UUID")
	errInvalidLenFor2UUIDs = errors.New("invalid input length for 2 UUID, excepted 32 byte")
	errRevokedPAT          = errors.New("revoked pat")
	errCreatePAT           = errors.New("failed to create PAT")
	errUpdatePAT           = errors.New("failed to update PAT")
	errRetrievePAT         = errors.New("failed to retrieve PAT")
	errDeletePAT           = errors.New("failed to delete PAT")
	errRevokePAT           = errors.New("failed to revoke PAT")
	errClearAllScope       = errors.New("failed to clear all entry in scope")
)

type Service struct {
	pats   PATSRepository
	hasher Hasher
}

var _ Service = (*Service)(nil)

// New instantiates the auth service implementation.
func New(pats PATSRepository, hasher Hasher) Service {
	return &Service{
		pats:               pats,
		hasher:             hasher,
	}
}

func (svc Service) CreatePAT(ctx context.Context, token, name, description string, duration time.Duration, scope Scope) (PAT, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return PAT{}, err
	}

	id, err := svc.idProvider.ID()
	if err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	secret, hash, err := svc.generateSecretAndHash(key.User, id)
	if err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	now := time.Now()
	pat := PAT{
		ID:          id,
		User:        key.User,
		Name:        name,
		Description: description,
		Secret:      hash,
		IssuedAt:    now,
		ExpiresAt:   now.Add(duration),
		Scope:       scope,
	}
	if err := svc.pats.Save(ctx, pat); err != nil {
		return PAT{}, errors.Wrap(errCreatePAT, err)
	}
	pat.Secret = secret
	return pat, nil
}

func (svc Service) UpdatePATName(ctx context.Context, token, patID, name string) (PAT, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return PAT{}, err
	}
	pat, err := svc.pats.UpdateName(ctx, key.User, patID, name)
	if err != nil {
		return PAT{}, errors.Wrap(errUpdatePAT, err)
	}
	return pat, nil
}

func (svc Service) UpdatePATDescription(ctx context.Context, token, patID, description string) (PAT, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return PAT{}, err
	}
	pat, err := svc.pats.UpdateDescription(ctx, key.User, patID, description)
	if err != nil {
		return PAT{}, errors.Wrap(errUpdatePAT, err)
	}
	return pat, nil
}

func (svc Service) RetrievePAT(ctx context.Context, token, patID string) (PAT, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return PAT{}, err
	}

	pat, err := svc.pats.Retrieve(ctx, key.User, patID)
	if err != nil {
		return PAT{}, errors.Wrap(errRetrievePAT, err)
	}
	return pat, nil
}

func (svc Service) ListPATS(ctx context.Context, token string, pm PATSPageMeta) (PATSPage, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return PATSPage{}, err
	}
	patsPage, err := svc.pats.RetrieveAll(ctx, key.User, pm)
	if err != nil {
		return PATSPage{}, errors.Wrap(errRetrievePAT, err)
	}
	return patsPage, nil
}

func (svc Service) DeletePAT(ctx context.Context, token, patID string) error {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return err
	}
	if err := svc.pats.Remove(ctx, key.User, patID); err != nil {
		return errors.Wrap(errDeletePAT, err)
	}
	return nil
}

func (svc Service) ResetPATSecret(ctx context.Context, token, patID string, duration time.Duration) (PAT, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return PAT{}, err
	}

	// Generate new HashToken take place here
	secret, hash, err := svc.generateSecretAndHash(key.User, patID)
	if err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	pat, err := svc.pats.UpdateTokenHash(ctx, key.User, patID, hash, time.Now().Add(duration))
	if err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	if err := svc.pats.Reactivate(ctx, key.User, patID); err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	pat.Secret = secret
	pat.Revoked = false
	pat.RevokedAt = time.Time{}
	return pat, nil
}

func (svc Service) RevokePATSecret(ctx context.Context, token, patID string) error {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return err
	}

	if err := svc.pats.Revoke(ctx, key.User, patID); err != nil {
		return errors.Wrap(errRevokePAT, err)
	}
	return nil
}

func (svc Service) AddPATScopeEntry(ctx context.Context, token, patID string, platformEntityType PlatformEntityType, optionalDomainID string, optionalDomainEntityType DomainEntityType, operation OperationType, entityIDs ...string) (Scope, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return Scope{}, err
	}
	scope, err := svc.pats.AddScopeEntry(ctx, key.User, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	if err != nil {
		return Scope{}, errors.Wrap(errRevokePAT, err)
	}
	return scope, nil
}

func (svc Service) RemovePATScopeEntry(ctx context.Context, token, patID string, platformEntityType PlatformEntityType, optionalDomainID string, optionalDomainEntityType DomainEntityType, operation OperationType, entityIDs ...string) (Scope, error) {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return Scope{}, err
	}
	scope, err := svc.pats.RemoveScopeEntry(ctx, key.User, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	if err != nil {
		return Scope{}, err
	}
	return scope, nil
}

func (svc Service) ClearPATAllScopeEntry(ctx context.Context, token, patID string) error {
	key, err := svc.Identify(ctx, token)
	if err != nil {
		return err
	}
	if err := svc.pats.RemoveAllScopeEntry(ctx, key.User, patID); err != nil {
		return errors.Wrap(errClearAllScope, err)
	}
	return nil
}

func (svc Service) IdentifyPAT(ctx context.Context, secret string) (PAT, error) {
	parts := strings.Split(secret, patSecretSeparator)
	if len(parts) != 3 && parts[0] != patPrefix {
		return PAT{}, errors.Wrap(svcerr.ErrAuthentication, errMalformedPAT)
	}
	userID, patID, err := decode(parts[1])
	if err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrAuthentication, errMalformedPAT)
	}
	secretHash, revoked, err := svc.pats.RetrieveSecretAndRevokeStatus(ctx, userID.String(), patID.String())
	if err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if revoked {
		return PAT{}, errors.Wrap(svcerr.ErrAuthentication, errRevokedPAT)
	}
	if err := svc.hasher.Compare(secret, secretHash); err != nil {
		return PAT{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	return PAT{ID: patID.String(), User: userID.String()}, nil
}

func (svc Service) AuthorizePAT(ctx context.Context, paToken string, platformEntityType PlatformEntityType, optionalDomainID string, optionalDomainEntityType DomainEntityType, operation OperationType, entityIDs ...string) error {
	res, err := svc.IdentifyPAT(ctx, paToken)
	if err != nil {
		return err
	}
	if err := svc.pats.CheckScopeEntry(ctx, res.User, res.ID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...); err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}

func (svc Service) CheckPAT(ctx context.Context, userID, patID string, platformEntityType PlatformEntityType, optionalDomainID string, optionalDomainEntityType DomainEntityType, operation OperationType, entityIDs ...string) error {
	if err := svc.pats.CheckScopeEntry(ctx, userID, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...); err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}

func (svc Service) generateSecretAndHash(userID, patID string) (string, string, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return "", "", errors.Wrap(errFailedToParseUUID, err)
	}
	pID, err := uuid.Parse(patID)
	if err != nil {
		return "", "", errors.Wrap(errFailedToParseUUID, err)
	}

	secret := patPrefix + patSecretSeparator + encode(uID, pID) + patSecretSeparator + generateRandomString(100)
	secretHash, err := svc.hasher.Hash(secret)
	return secret, secretHash, err
}

func encode(userID, patID uuid.UUID) string {
	c := append(userID[:], patID[:]...)
	return base64.StdEncoding.EncodeToString(c)
}

func decode(encoded string) (uuid.UUID, uuid.UUID, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	if len(data) != 32 {
		return uuid.Nil, uuid.Nil, errInvalidLenFor2UUIDs
	}

	var userID, patID uuid.UUID
	copy(userID[:], data[:16])
	copy(patID[:], data[16:])

	return userID, patID, nil
}

func generateRandomString(n int) string {
	letterRunes := []rune(randStr)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
