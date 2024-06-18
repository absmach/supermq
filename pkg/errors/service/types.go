// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package service

import "github.com/absmach/magistrala/pkg/errors"

// Wrapper for Service errors.
// ErrAuthentication indicates failure occurred while authenticating the entity.
type AuthenticationError struct {
	Err errors.Error
}

func (ae *AuthenticationError) Error() string {
	return ae.Err.Error()
}

var ErrAuthentication = &AuthenticationError{
	Err: errors.New("failed to perform authentication over the entity"),
}

// ErrAuthorization indicates failure occurred while authorizing the entity.
type AuthorizationError struct {
	Err errors.Error
}

func (ae *AuthorizationError) Error() string {
	return ae.Err.Error()
}

var ErrAuthorization = &AuthorizationError{
	Err: errors.New("failed to perform authorization over the entity"),
}

// ErrDomainAuthorization indicates failure occurred while authorizing the domain.
type DomainAuthorizationError struct {
	Err errors.Error
}

func (dae *DomainAuthorizationError) Error() string {
	return dae.Err.Error()
}

var ErrDomainAuthorization = &DomainAuthorizationError{
	Err: errors.New("failed to perform authorization over the domain"),
}

// ErrLogin indicates wrong login credentials.
type LoginError struct {
	Err errors.Error
}

func (le *LoginError) Error() string {
	return le.Err.Error()
}

var ErrLogin = &LoginError{
	Err: errors.New("invalid user id or secret"),
}

// ErrMalformedEntity indicates a malformed entity specification.
type MalformedEntityError struct {
	Err errors.Error
}

func (mee *MalformedEntityError) Error() string {
	return mee.Err.Error()
}

var ErrMalformedEntity = &MalformedEntityError{
	Err: errors.New("malformed entity specification"),
}

// ErrNotFound indicates a non-existent entity request.
type NotFoundError struct {
	Err errors.Error
}

func (nfe *NotFoundError) Error() string {
	return nfe.Err.Error()
}

var ErrNotFound = &NotFoundError{
	Err: errors.New("entity not found"),
}

// ErrConflict indicates that entity already exists.
type ConflictError struct {
	Err errors.Error
}

func (ce *ConflictError) Error() string {
	return ce.Err.Error()
}

var ErrConflict = &ConflictError{
	Err: errors.New("entity already exists"),
}

// ErrCreateEntity indicates error in creating entity or entities.
type CreateEntityError struct {
	Err errors.Error
}

func (cee *CreateEntityError) Error() string {
	return cee.Err.Error()
}

var ErrCreateEntity = &CreateEntityError{
	Err: errors.New("failed to create entity"),
}

// ErrRemoveEntity indicates error in removing entity.
type RemoveEntityError struct {
	Err errors.Error
}

func (ree *RemoveEntityError) Error() string {
	return ree.Err.Error()
}

var ErrRemoveEntity = &RemoveEntityError{
	Err: errors.New("failed to remove entity"),
}

// ErrViewEntity indicates error in viewing entity or entities.
type ViewEntityError struct {
	Err errors.Error
}

func (vee *ViewEntityError) Error() string {
	return vee.Err.Error()
}

var ErrViewEntity = &ViewEntityError{
	Err: errors.New("view entity failed"),
}

// ErrUpdateEntity indicates error in updating entity or entities.
type UpdateEntityError struct {
	Err errors.Error
}

func (uee *UpdateEntityError) Error() string {
	return uee.Err.Error()
}

var ErrUpdateEntity = &UpdateEntityError{
	Err: errors.New("update entity failed"),
}

// ErrInvalidStatus indicates an invalid status.
type InvalidStatusError struct {
	Err errors.Error
}

func (ise *InvalidStatusError) Error() string {
	return ise.Err.Error()
}

var ErrInvalidStatus = &InvalidStatusError{
	Err: errors.New("invalid status"),
}

// ErrInvalidRole indicates that an invalid role.
type InvalidRoleError struct {
	Err errors.Error
}

func (ire *InvalidRoleError) Error() string {
	return ire.Err.Error()
}

var ErrInvalidRole = &InvalidRoleError{
	Err: errors.New("invalid client role"),
}

// ErrInvalidPolicy indicates that an invalid policy.
type InvalidPolicyError struct {
	Err errors.Error
}

func (ipe *InvalidPolicyError) Error() string {
	return ipe.Err.Error()
}

var ErrInvalidPolicy = &InvalidPolicyError{
	Err: errors.New("invalid policy"),
}

// ErrEnableClient indicates error in enabling client.
type EnableClientError struct {
	Err errors.Error
}

func (ece *EnableClientError) Error() string {
	return ece.Err.Error()
}

var ErrEnableClient = &EnableClientError{
	Err: errors.New("failed to enable client"),
}

// ErrDisableClient indicates error in disabling client.
type DisableClientError struct {
	Err errors.Error
}

func (dce *DisableClientError) Error() string {
	return dce.Err.Error()
}

var ErrDisableClient = &DisableClientError{
	Err: errors.New("failed to disable client"),
}

// ErrAddPolicies indicates error in adding policies.
type AddPoliciesError struct {
	Err errors.Error
}

func (ape *AddPoliciesError) Error() string {
	return ape.Err.Error()
}

var ErrAddPolicies = &AddPoliciesError{
	Err: errors.New("failed to add policies"),
}

// ErrDeletePolicies indicates error in removing policies.
type DeletePoliciesError struct {
	Err errors.Error
}

func (dpe *DeletePoliciesError) Error() string {
	return dpe.Err.Error()
}

var ErrDeletePolicies = &DeletePoliciesError{
	Err: errors.New("failed to remove policies"),
}
