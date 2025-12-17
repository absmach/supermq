// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package service

import "github.com/absmach/supermq/pkg/errors"

// Wrapper for Service errors.
var (
	// ErrAuthentication indicates failure occurred while authenticating the entity.
	ErrAuthentication = errors.NewAuthNError("failed to perform authentication over the entity")

	// ErrLogin indicates wrong login credentials.
	ErrLogin = errors.NewAuthNError("invalid credentials")

	// ErrAuthorization indicates failure occurred while authorizing the entity.
	ErrAuthorization = errors.NewAuthZError("failed to perform authorization over the entity")

	// ErrDomainAuthorization indicates failure occurred while authorizing the domain.
	ErrDomainAuthorization = errors.NewAuthZError("failed to perform authorization over the domain")

	// ErrUnauthorizedPAT indicates failure occurred while authorizing PAT.
	ErrUnauthorizedPAT = errors.NewAuthZError("failed to authorize PAT")

	// ErrSuperAdminAction indicates that the user is not a super admin.
	ErrSuperAdminAction = errors.NewAuthZError("not authorized to perform admin action")

	// ErrCreateEntity indicates error in creating entity or entities.
	ErrCreateEntity = errors.NewServiceError("failed to create entity")

	// ErrRemoveEntity indicates error in removing entity.
	ErrRemoveEntity = errors.NewServiceError("failed to remove entity")

	// ErrViewEntity indicates error in viewing entity or entities.
	ErrViewEntity = errors.NewServiceError("view entity failed")

	// ErrUpdateEntity indicates error in updating entity or entities.
	ErrUpdateEntity = errors.NewServiceError("update entity failed")
	
	// ErrAddPolicies indicates error in adding policies.
	ErrAddPolicies = errors.NewServiceError("failed to add policies")

	// ErrUserAlreadyVerified indicates user is already verified.
	ErrUserAlreadyVerified = errors.NewServiceError("user already verified")

	// ErrInvalidUserVerification indicates user verification is invalid.
	ErrInvalidUserVerification = errors.NewServiceError("invalid verification")

	// ErrIssueProviderID indicates failure to issue unique ID from ID provider.
	ErrIssueProviderID = errors.NewServiceError("failed to issue unique ID from id provider")

	// ErrHashPassword indicates failure to hash password.
	ErrHashPassword = errors.NewServiceError("failed to hash password")

	// Request Errors (NewRequestError)
	// ErrMissingUsername indicates that the user's names are missing.
	ErrMissingUsername = errors.NewRequestError("missing usernames")

	// ErrMalformedEntity indicates a malformed entity specification.
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("entity not found")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.New("entity already exists")

	// ErrInvalidStatus indicates an invalid status.
	ErrInvalidStatus = errors.New("invalid status")

	// ErrInvalidRole indicates that an invalid role.
	ErrInvalidRole = errors.New("invalid client role")

	// ErrInvalidPolicy indicates that an invalid policy.
	ErrInvalidPolicy = errors.New("invalid policy")

	// ErrEnableClient indicates error in enabling client.
	ErrEnableClient = errors.New("failed to enable client")

	// ErrDisableClient indicates error in disabling client.
	ErrDisableClient = errors.New("failed to disable client")

	// ErrDeletePolicies indicates error in removing policies.
	ErrDeletePolicies = errors.New("failed to remove policies")

	// ErrSearch indicates error in searching clients.
	ErrSearch = errors.New("failed to search clients")

	// ErrInvitationAlreadyRejected indicates that the invitation is already rejected.
	ErrInvitationAlreadyRejected = errors.New("invitation already rejected")

	// ErrInvitationAlreadyAccepted indicates that the invitation is already accepted.
	ErrInvitationAlreadyAccepted = errors.New("invitation already accepted")

	// ErrParentGroupAuthorization indicates failure occurred while authorizing the parent group.
	ErrParentGroupAuthorization = errors.New("failed to authorize parent group")

	// ErrEnableUser indicates error in enabling user.
	ErrEnableUser = errors.New("failed to enable user")

	// ErrDisableUser indicates error in disabling user.
	ErrDisableUser = errors.New("failed to disable user")

	// ErrRollbackRepo indicates a failure to rollback repository.
	ErrRollbackRepo = errors.New("failed to rollback repo")

	// ErrRetainOneMember indicates that at least one owner must be retained in the entity.
	ErrRetainOneMember = errors.New("must retain at least one member")

	// ErrUserVerificationExpired indicates user verification is expired.
	ErrUserVerificationExpired = errors.New("verification expired, please generate New verification")

	// ErrRegisterUser indicates error in register a user.
	ErrRegisterUser = errors.New("failed to register user")

	// ErrExternalAuthProviderCouldNotUpdate indicates that users authenticated via external provider cannot update their account details directly.
	ErrExternalAuthProviderCouldNotUpdate = errors.New("account details can only be updated through your authentication provider's settings")

	// ErrFailedToSaveEntityDB indicates failure to save entity to database.
	ErrFailedToSaveEntityDB = errors.New("failed to save entity to database")
)
