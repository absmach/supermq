// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package service

import "github.com/absmach/supermq/pkg/errors"

// Wrapper for Service errors.
var (
	// ErrAuthentication indicates failure occurred while authenticating the entity.
	ErrAuthentication = errors.NewServiceError("failed to perform authentication over the entity")

	// ErrAuthorization indicates failure occurred while authorizing the entity.
	ErrAuthorization = errors.NewServiceError("failed to perform authorization over the entity")

	// ErrDomainAuthorization indicates failure occurred while authorizing the domain.
	ErrDomainAuthorization = errors.NewServiceError("failed to perform authorization over the domain")

	// ErrLogin indicates wrong login credentials.
	ErrLogin = errors.NewServiceError("invalid credentials")

	// ErrMalformedEntity indicates a malformed entity specification.
	ErrMalformedEntity = errors.NewServiceError("malformed entity specification")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.NewServiceError("entity not found")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.NewServiceError("entity already exists")

	// ErrCreateEntity indicates error in creating entity or entities.
	ErrCreateEntity = errors.NewServiceError("failed to create entity")

	// ErrRemoveEntity indicates error in removing entity.
	ErrRemoveEntity = errors.NewServiceError("failed to remove entity")

	// ErrViewEntity indicates error in viewing entity or entities.
	ErrViewEntity = errors.NewServiceError("view entity failed")

	// ErrUpdateEntity indicates error in updating entity or entities.
	ErrUpdateEntity = errors.NewServiceError("update entity failed")

	// ErrInvalidStatus indicates an invalid status.
	ErrInvalidStatus = errors.NewServiceError("invalid status")

	// ErrInvalidRole indicates that an invalid role.
	ErrInvalidRole = errors.NewServiceError("invalid client role")

	// ErrInvalidPolicy indicates that an invalid policy.
	ErrInvalidPolicy = errors.NewServiceError("invalid policy")

	// ErrEnableClient indicates error in enabling client.
	ErrEnableClient = errors.NewServiceError("failed to enable client")

	// ErrDisableClient indicates error in disabling client.
	ErrDisableClient = errors.NewServiceError("failed to disable client")

	// ErrAddPolicies indicates error in adding policies.
	ErrAddPolicies = errors.NewServiceError("failed to add policies")

	// ErrDeletePolicies indicates error in removing policies.
	ErrDeletePolicies = errors.NewServiceError("failed to remove policies")

	// ErrSearch indicates error in searching clients.
	ErrSearch = errors.NewServiceError("failed to search clients")

	// ErrInvitationAlreadyRejected indicates that the invitation is already rejected.
	ErrInvitationAlreadyRejected = errors.NewServiceError("invitation already rejected")

	// ErrInvitationAlreadyAccepted indicates that the invitation is already accepted.
	ErrInvitationAlreadyAccepted = errors.NewServiceError("invitation already accepted")

	// ErrParentGroupAuthorization indicates failure occurred while authorizing the parent group.
	ErrParentGroupAuthorization = errors.NewServiceError("failed to authorize parent group")

	// ErrMissingUsername indicates that the user's names are missing.
	ErrMissingUsername = errors.NewServiceError("missing usernames")

	// ErrEnableUser indicates error in enabling user.
	ErrEnableUser = errors.NewServiceError("failed to enable user")

	// ErrDisableUser indicates error in disabling user.
	ErrDisableUser = errors.NewServiceError("failed to disable user")

	// ErrRollbackRepo indicates a failure to rollback repository.
	ErrRollbackRepo = errors.NewServiceError("failed to rollback repo")

	// ErrUnauthorizedPAT indicates failure occurred while authorizing PAT.
	ErrUnauthorizedPAT = errors.NewServiceError("failed to authorize PAT")

	// ErrRetainOneMember indicates that at least one owner must be retained in the entity.
	ErrRetainOneMember = errors.NewServiceError("must retain at least one member")

	// ErrSuperAdminAction indicates that the user is not a super admin.
	ErrSuperAdminAction = errors.NewServiceError("not authorized to perform admin action")

	// ErrUserAlreadyVerified indicates user is already verified.
	ErrUserAlreadyVerified = errors.NewServiceError("user already verified")

	// ErrInvalidUserVerification indicates user verification is invalid.
	ErrInvalidUserVerification = errors.NewServiceError("invalid verification")

	// ErrUserVerificationExpired indicates user verification is expired.
	ErrUserVerificationExpired = errors.NewServiceError("verification expired, please generate NewServiceError verification")

	// ErrRegisterUser indicates error in register a user.
	ErrRegisterUser = errors.NewServiceError("failed to register user")

	// ErrExternalAuthProviderCouldNotUpdate indicates that users authenticated via external provider cannot update their account details directly.
	ErrExternalAuthProviderCouldNotUpdate = errors.NewServiceError("account details can only be updated through your authentication provider's settings")

	// ErrFailedToSaveEntityDB indicates failure to save entity to database.
	ErrFailedToSaveEntityDB = errors.NewServiceError("failed to save entity to database")

	// ErrIssueProviderID indicates failure to issue unique ID from ID provider.
	ErrIssueProviderID = errors.NewServiceError("failed to issue unique ID from id provider")

	// ErrHashPassword indicates failure to hash password.
	ErrHashPassword = errors.NewServiceError("failed to hash password")
)
