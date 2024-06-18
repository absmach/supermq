// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

// ErrMalformedEntity indicates a malformed entity specification.
type MalformedEntityError struct {
	Err Error
}

func (mee *MalformedEntityError) Error() string {
	return mee.Err.Error()
}

var ErrMalformedEntity = &MalformedEntityError{
	Err: New("malformed entity specification"),
}

// ErrUnsupportedContentType indicates invalid content type.
type UnsupportedContentTypeError struct {
	Err Error
}

func (ucte *UnsupportedContentTypeError) Error() string {
	return ucte.Err.Error()
}

var ErrUnsupportedContentType = &UnsupportedContentTypeError{
	Err: New("invalid content type"),
}

// ErrUnidentified indicates unidentified error.
type UnidentifiedError struct {
	Err Error
}

func (ue *UnidentifiedError) Error() string {
	return ue.Err.Error()
}

var ErrUnidentified = &UnidentifiedError{
	Err: New("unidentified error"),
}

// ErrEmptyPath indicates empty file path.
type EmptyPathError struct {
	Err Error
}

func (epe *EmptyPathError) Error() string {
	return epe.Err.Error()
}

var ErrEmptyPath = &EmptyPathError{
	Err: New("empty file path"),
}

// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
type StatusAlreadyAssignedError struct {
	Err Error
}

func (sae *StatusAlreadyAssignedError) Error() string {
	return sae.Err.Error()
}

var ErrStatusAlreadyAssigned = &StatusAlreadyAssignedError{
	Err: New("status already assigned"),
}

// ErrRollbackTx indicates failed to rollback transaction.

type RollbackTxError struct {
	Err Error
}

func (rte *RollbackTxError) Error() string {
	return rte.Err.Error()
}

var ErrRollbackTx = &RollbackTxError{
	Err: New("failed to rollback transaction"),
}
