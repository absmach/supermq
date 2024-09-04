// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"strings"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/domains"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
)

const (
	recoveryDuration = 5 * time.Minute
	defLimit         = 100
)

var (
	// ErrExpiry indicates that the token is expired.
	ErrExpiry = errors.New("token is expired")

	errIssueUser          = errors.New("failed to issue new login key")
	errIssueTmp           = errors.New("failed to issue new temporary key")
	errRevoke             = errors.New("failed to remove key")
	errRetrieve           = errors.New("failed to retrieve key data")
	errIdentify           = errors.New("failed to validate token")
	errPlatform           = errors.New("invalid platform id")
	errCreateDomainPolicy = errors.New("failed to create domain policy")
	errAddPolicies        = errors.New("failed to add policies")
	errRemovePolicies     = errors.New("failed to remove the policies")
	errRollbackPolicy     = errors.New("failed to rollback policy")
	errRemoveLocalPolicy  = errors.New("failed to remove from local policy copy")
	errRemovePolicyEngine = errors.New("failed to remove from policy engine")
	// errInvalidEntityType indicates invalid entity type.
	errInvalidEntityType = errors.New("invalid entity type")
)

var (
	defThingsFilterPermissions = []string{
		AdminPermission,
		DeletePermission,
		EditPermission,
		ViewPermission,
		SharePermission,
		PublishPermission,
		SubscribePermission,
	}

	defGroupsFilterPermissions = []string{
		AdminPermission,
		DeletePermission,
		EditPermission,
		ViewPermission,
		MembershipPermission,
		SharePermission,
	}

	defDomainsFilterPermissions = []string{
		AdminPermission,
		EditPermission,
		ViewPermission,
		MembershipPermission,
		SharePermission,
	}

	defPlatformFilterPermissions = []string{
		AdminPermission,
		MembershipPermission,
	}
)

// Authn specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
// Token is a string value of the actual Key and is used to authenticate
// an Auth service request.
type Authn interface {
	// Issue issues a new Key, returning its token value alongside.
	Issue(ctx context.Context, token string, key Key) (Token, error)

	// Revoke removes the Key with the provided id that is
	// issued by the user identified by the provided key.
	Revoke(ctx context.Context, token, id string) error

	// RetrieveKey retrieves data for the Key identified by the provided
	// ID, that is issued by the user identified by the provided key.
	RetrieveKey(ctx context.Context, token, id string) (Key, error)

	// Identify validates token token. If token is valid, content
	// is returned. If token is invalid, or invocation failed for some
	// other reason, non-nil error value is returned in response.
	Identify(ctx context.Context, token string) (Key, error)
}

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
// Token is a string value of the actual Key and is used to authenticate
// an Auth service request.

//go:generate mockery --name Service --output=./mocks --filename service.go --quiet --note "Copyright (c) Abstract Machines"
type Service interface {
	Authn
	Authz
}

var _ Service = (*service)(nil)

type service struct {
	keys               KeyRepository
	idProvider         magistrala.IDProvider
	agent              PolicyAgent
	tokenizer          Tokenizer
	loginDuration      time.Duration
	refreshDuration    time.Duration
	invitationDuration time.Duration
}

// New instantiates the auth service implementation.
func New(keys KeyRepository, idp magistrala.IDProvider, tokenizer Tokenizer, policyAgent PolicyAgent, loginDuration, refreshDuration, invitationDuration time.Duration) Service {
	return &service{
		tokenizer:          tokenizer,
		keys:               keys,
		idProvider:         idp,
		agent:              policyAgent,
		loginDuration:      loginDuration,
		refreshDuration:    refreshDuration,
		invitationDuration: invitationDuration,
	}
}

func (svc service) Issue(ctx context.Context, token string, key Key) (Token, error) {
	key.IssuedAt = time.Now().UTC()
	switch key.Type {
	case APIKey:
		return svc.userKey(ctx, token, key)
	case RefreshKey:
		return svc.refreshKey(ctx, token, key)
	case RecoveryKey:
		return svc.tmpKey(recoveryDuration, key)
	case InvitationKey:
		return svc.invitationKey(ctx, key)
	default:
		return svc.accessKey(ctx, key)
	}
}

func (svc service) Revoke(ctx context.Context, token, id string) error {
	issuerID, _, err := svc.authenticate(token)
	if err != nil {
		return errors.Wrap(errRevoke, err)
	}
	if err := svc.keys.Remove(ctx, issuerID, id); err != nil {
		return errors.Wrap(errRevoke, err)
	}
	return nil
}

func (svc service) RetrieveKey(ctx context.Context, token, id string) (Key, error) {
	issuerID, _, err := svc.authenticate(token)
	if err != nil {
		return Key{}, errors.Wrap(errRetrieve, err)
	}

	key, err := svc.keys.Retrieve(ctx, issuerID, id)
	if err != nil {
		return Key{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return key, nil
}

func (svc service) Identify(ctx context.Context, token string) (Key, error) {
	key, err := svc.tokenizer.Parse(token)
	if errors.Contains(err, ErrExpiry) {
		err = svc.keys.Remove(ctx, key.Issuer, key.ID)
		return Key{}, errors.Wrap(svcerr.ErrAuthentication, errors.Wrap(ErrKeyExpired, err))
	}
	if err != nil {
		return Key{}, errors.Wrap(svcerr.ErrAuthentication, errors.Wrap(errIdentify, err))
	}

	switch key.Type {
	case RecoveryKey, AccessKey, InvitationKey, RefreshKey:
		return key, nil
	case APIKey:
		_, err := svc.keys.Retrieve(ctx, key.Issuer, key.ID)
		if err != nil {
			return Key{}, svcerr.ErrAuthentication
		}
		return key, nil
	default:
		return Key{}, svcerr.ErrAuthentication
	}
}

func (svc service) Authorize(ctx context.Context, pr PolicyReq) error {
	if err := svc.PolicyValidation(pr); err != nil {
		return errors.Wrap(svcerr.ErrMalformedEntity, err)
	}
	if pr.SubjectKind == TokenKind {
		key, err := svc.Identify(ctx, pr.Subject)
		if err != nil {
			return errors.Wrap(svcerr.ErrAuthentication, err)
		}
		if key.Subject == "" {
			if pr.ObjectType == GroupType || pr.ObjectType == ThingType || pr.ObjectType == DomainType {
				return svcerr.ErrDomainAuthorization
			}
			return svcerr.ErrAuthentication
		}
		pr.Subject = key.Subject
		pr.Domain = key.Domain
	}
	if err := svc.checkPolicy(ctx, pr); err != nil {
		return err
	}
	return nil
}

func (svc service) checkPolicy(ctx context.Context, pr PolicyReq) error {
	// Domain status is required for if user sent authorization request on things, channels, groups and domains
	if pr.SubjectType == UserType && (pr.ObjectType == GroupType || pr.ObjectType == ThingType || pr.ObjectType == DomainType) {
		domainID := pr.Domain
		if domainID == "" {
			if pr.ObjectType != DomainType {
				return svcerr.ErrDomainAuthorization
			}
			domainID = pr.Object
		}
		if err := svc.checkDomain(ctx, pr.SubjectType, pr.Subject, domainID); err != nil {
			return err
		}
	}
	if err := svc.agent.CheckPolicy(ctx, pr); err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}

func (svc service) checkDomain(ctx context.Context, subjectType, subject, domainID string) error {
	if err := svc.agent.CheckPolicy(ctx, PolicyReq{
		Subject:     subject,
		SubjectType: subjectType,
		Permission:  domains.MembershipPermission,
		Object:      domainID,
		ObjectType:  DomainType,
	}); err != nil {
		return svcerr.ErrDomainAuthorization
	}

	// ToDo: Add domain status in spiceDB like with new relation called status

	// d, err := svc.domains.RetrieveByID(ctx, domainID)
	// if err != nil {
	// 	return errors.Wrap(svcerr.ErrViewEntity, err)
	// }

	// switch d.Status {
	// case EnabledStatus:
	// case DisabledStatus:
	// 	if err := svc.agent.CheckPolicy(ctx, PolicyReq{
	// 		Subject:     subject,
	// 		SubjectType: subjectType,
	// 		Permission:  AdminPermission,
	// 		Object:      domainID,
	// 		ObjectType:  DomainType,
	// 	}); err != nil {
	// 		return svcerr.ErrDomainAuthorization
	// 	}
	// case FreezeStatus:
	// 	if err := svc.agent.CheckPolicy(ctx, PolicyReq{
	// 		Subject:     subject,
	// 		SubjectType: subjectType,
	// 		Permission:  AdminPermission,
	// 		Object:      MagistralaObject,
	// 		ObjectType:  PlatformType,
	// 	}); err != nil {
	// 		return svcerr.ErrDomainAuthorization
	// 	}
	// default:
	// 	return svcerr.ErrDomainAuthorization
	// }

	return nil
}

func (svc service) AddPolicy(ctx context.Context, pr PolicyReq) error {
	if err := svc.PolicyValidation(pr); err != nil {
		return errors.Wrap(svcerr.ErrInvalidPolicy, err)
	}
	return svc.agent.AddPolicy(ctx, pr)
}

func (svc service) PolicyValidation(pr PolicyReq) error {
	if pr.ObjectType == PlatformType && pr.Object != MagistralaObject {
		return errPlatform
	}
	return nil
}

func (svc service) AddPolicies(ctx context.Context, prs []PolicyReq) error {
	for _, pr := range prs {
		if err := svc.PolicyValidation(pr); err != nil {
			return errors.Wrap(svcerr.ErrInvalidPolicy, err)
		}
	}
	return svc.agent.AddPolicies(ctx, prs)
}

func (svc service) DeletePolicyFilter(ctx context.Context, pr PolicyReq) error {
	return svc.agent.DeletePolicyFilter(ctx, pr)
}

func (svc service) DeletePolicies(ctx context.Context, prs []PolicyReq) error {
	for _, pr := range prs {
		if err := svc.PolicyValidation(pr); err != nil {
			return errors.Wrap(svcerr.ErrInvalidPolicy, err)
		}
	}
	return svc.agent.DeletePolicies(ctx, prs)
}

func (svc service) ListObjects(ctx context.Context, pr PolicyReq, nextPageToken string, limit uint64) (PolicyPage, error) {
	if limit <= 0 {
		limit = 100
	}
	res, npt, err := svc.agent.RetrieveObjects(ctx, pr, nextPageToken, limit)
	if err != nil {
		return PolicyPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	var page PolicyPage
	for _, tuple := range res {
		page.Policies = append(page.Policies, tuple.Object)
	}
	page.NextPageToken = npt
	return page, nil
}

func (svc service) ListAllObjects(ctx context.Context, pr PolicyReq) (PolicyPage, error) {
	res, err := svc.agent.RetrieveAllObjects(ctx, pr)
	if err != nil {
		return PolicyPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	var page PolicyPage
	for _, tuple := range res {
		page.Policies = append(page.Policies, tuple.Object)
	}
	return page, nil
}

func (svc service) CountObjects(ctx context.Context, pr PolicyReq) (uint64, error) {
	return svc.agent.RetrieveAllObjectsCount(ctx, pr)
}

func (svc service) ListSubjects(ctx context.Context, pr PolicyReq, nextPageToken string, limit uint64) (PolicyPage, error) {
	if limit <= 0 {
		limit = 100
	}
	res, npt, err := svc.agent.RetrieveSubjects(ctx, pr, nextPageToken, limit)
	if err != nil {
		return PolicyPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	var page PolicyPage
	for _, tuple := range res {
		page.Policies = append(page.Policies, tuple.Subject)
	}
	page.NextPageToken = npt
	return page, nil
}

func (svc service) ListAllSubjects(ctx context.Context, pr PolicyReq) (PolicyPage, error) {
	res, err := svc.agent.RetrieveAllSubjects(ctx, pr)
	if err != nil {
		return PolicyPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	var page PolicyPage
	for _, tuple := range res {
		page.Policies = append(page.Policies, tuple.Subject)
	}
	return page, nil
}

func (svc service) CountSubjects(ctx context.Context, pr PolicyReq) (uint64, error) {
	return svc.agent.RetrieveAllSubjectsCount(ctx, pr)
}

func (svc service) ListPermissions(ctx context.Context, pr PolicyReq, permissionsFilter []string) (Permissions, error) {
	if len(permissionsFilter) == 0 {
		switch pr.ObjectType {
		case ThingType:
			permissionsFilter = defThingsFilterPermissions
		case GroupType:
			permissionsFilter = defGroupsFilterPermissions
		case PlatformType:
			permissionsFilter = defPlatformFilterPermissions
		case DomainType:
			permissionsFilter = defDomainsFilterPermissions
		default:
			return nil, svcerr.ErrMalformedEntity
		}
	}
	pers, err := svc.agent.RetrievePermissions(ctx, pr, permissionsFilter)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return pers, nil
}

func (svc service) tmpKey(duration time.Duration, key Key) (Token, error) {
	key.ExpiresAt = time.Now().Add(duration)
	value, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Token{}, errors.Wrap(errIssueTmp, err)
	}

	return Token{AccessToken: value}, nil
}

func (svc service) accessKey(ctx context.Context, key Key) (Token, error) {
	var err error
	key.Type = AccessKey
	key.ExpiresAt = time.Now().Add(svc.loginDuration)

	key.Subject, err = svc.checkUserDomain(ctx, key)
	if err != nil {
		return Token{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}

	access, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Token{}, errors.Wrap(errIssueTmp, err)
	}

	key.ExpiresAt = time.Now().Add(svc.refreshDuration)
	key.Type = RefreshKey
	refresh, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Token{}, errors.Wrap(errIssueTmp, err)
	}

	return Token{AccessToken: access, RefreshToken: refresh}, nil
}

func (svc service) invitationKey(ctx context.Context, key Key) (Token, error) {
	var err error
	key.Type = InvitationKey
	key.ExpiresAt = time.Now().Add(svc.invitationDuration)

	key.Subject, err = svc.checkUserDomain(ctx, key)
	if err != nil {
		return Token{}, err
	}

	access, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Token{}, errors.Wrap(errIssueTmp, err)
	}

	return Token{AccessToken: access}, nil
}

func (svc service) refreshKey(ctx context.Context, token string, key Key) (Token, error) {
	k, err := svc.tokenizer.Parse(token)
	if err != nil {
		return Token{}, errors.Wrap(errRetrieve, err)
	}
	if k.Type != RefreshKey {
		return Token{}, errIssueUser
	}
	key.ID = k.ID
	if key.Domain == "" {
		key.Domain = k.Domain
	}
	key.User = k.User
	key.Type = AccessKey

	key.Subject, err = svc.checkUserDomain(ctx, key)
	if err != nil {
		return Token{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}

	key.ExpiresAt = time.Now().Add(svc.loginDuration)
	access, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Token{}, errors.Wrap(errIssueTmp, err)
	}

	key.ExpiresAt = time.Now().Add(svc.refreshDuration)
	key.Type = RefreshKey
	refresh, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Token{}, errors.Wrap(errIssueTmp, err)
	}

	return Token{AccessToken: access, RefreshToken: refresh}, nil
}

func (svc service) checkUserDomain(ctx context.Context, key Key) (subject string, err error) {
	if key.Domain != "" {
		// Check user is platform admin.
		if err = svc.Authorize(ctx, PolicyReq{
			Subject:     key.User,
			SubjectType: UserType,
			Permission:  AdminPermission,
			Object:      MagistralaObject,
			ObjectType:  PlatformType,
		}); err == nil {
			return key.User, nil
		}
		// Check user is domain member.
		domainUserSubject := EncodeDomainUserID(key.Domain, key.User)
		if err = svc.Authorize(ctx, PolicyReq{
			Subject:     domainUserSubject,
			SubjectType: UserType,
			Permission:  MembershipPermission,
			Object:      key.Domain,
			ObjectType:  DomainType,
		}); err != nil {
			return "", err
		}
		return domainUserSubject, nil
	}
	return "", nil
}

func (svc service) userKey(ctx context.Context, token string, key Key) (Token, error) {
	id, sub, err := svc.authenticate(token)
	if err != nil {
		return Token{}, errors.Wrap(errIssueUser, err)
	}

	key.Issuer = id
	if key.Subject == "" {
		key.Subject = sub
	}

	keyID, err := svc.idProvider.ID()
	if err != nil {
		return Token{}, errors.Wrap(errIssueUser, err)
	}
	key.ID = keyID

	if _, err := svc.keys.Save(ctx, key); err != nil {
		return Token{}, errors.Wrap(errIssueUser, err)
	}

	tkn, err := svc.tokenizer.Issue(key)
	if err != nil {
		return Token{}, errors.Wrap(errIssueUser, err)
	}

	return Token{AccessToken: tkn}, nil
}

func (svc service) authenticate(token string) (string, string, error) {
	key, err := svc.tokenizer.Parse(token)
	if err != nil {
		return "", "", errors.Wrap(svcerr.ErrAuthentication, err)
	}
	// Only login key token is valid for login.
	if key.Type != AccessKey || key.Issuer == "" {
		return "", "", svcerr.ErrAuthentication
	}

	return key.Issuer, key.Subject, nil
}

// Switch the relative permission for the relation.
func SwitchToPermission(relation string) string {
	switch relation {
	case AdministratorRelation:
		return AdminPermission
	case EditorRelation:
		return EditPermission
	case ContributorRelation:
		return ViewPermission
	case MemberRelation:
		return MembershipPermission
	case GuestRelation:
		return ViewPermission
	default:
		return relation
	}
}

func EncodeDomainUserID(domainID, userID string) string {
	if domainID == "" || userID == "" {
		return ""
	}
	return domainID + "_" + userID
}

func DecodeDomainUserID(domainUserID string) (string, string) {
	if domainUserID == "" {
		return domainUserID, domainUserID
	}
	duid := strings.Split(domainUserID, "_")

	switch {
	case len(duid) == 2:
		return duid[0], duid[1]
	case len(duid) == 1:
		return duid[0], ""
	case len(duid) == 0 || len(duid) > 2:
		fallthrough
	default:
		return "", ""
	}
}

func (svc service) DeleteEntityPolicies(ctx context.Context, entityType, id string) (err error) {
	switch entityType {
	case ThingType:
		req := PolicyReq{
			Object:     id,
			ObjectType: ThingType,
		}

		return svc.DeletePolicyFilter(ctx, req)
	case UserType:
		req := PolicyReq{
			Subject:     id,
			SubjectType: UserType,
		}
		if err := svc.agent.DeletePolicyFilter(ctx, req); err != nil {
			return err
		}

		return nil
	case GroupType:
		req := PolicyReq{
			SubjectType: GroupType,
			Subject:     id,
		}
		if err := svc.DeletePolicyFilter(ctx, req); err != nil {
			return err
		}

		req = PolicyReq{
			Object:     id,
			ObjectType: GroupType,
		}
		return svc.DeletePolicyFilter(ctx, req)
	default:
		return errInvalidEntityType
	}
}
