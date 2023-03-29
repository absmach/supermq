package clients

import "errors"

var (
	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid client status")

	// ErrEnableClient indicates error in enabling client.
	ErrEnableClient = errors.New("failed to enable client")

	// ErrDisableClient indicates error in disabling client.
	ErrDisableClient = errors.New("failed to disable client")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")

	// ErrMissingResetToken indicates malformed or missing reset token
	// for reseting password.
	ErrMissingResetToken = errors.New("missing reset token")

	// ErrRecoveryToken indicates error in generating password recovery token.
	ErrRecoveryToken = errors.New("failed to generate password recovery token")

	// ErrGetToken indicates error in getting signed token.
	ErrGetToken = errors.New("failed to fetch signed token")

	// ErrPasswordFormat indicates weak password.
	ErrPasswordFormat = errors.New("password does not meet the requirements")
)
