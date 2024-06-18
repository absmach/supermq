// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package apiutil

import "github.com/absmach/magistrala/pkg/errors"

// Errors defined in this file are used by the LoggingErrorEncoder decorator
// to distinguish and log API request validation errors and avoid that service
// errors are logged twice.

// ValidationError indicates that an error was returned by the API.
type ValidationError struct {
	Err errors.Error
}

func (eve *ValidationError) Error() string {
	return eve.Err.Error()
}

var ErrValidation = &ValidationError{
	Err: errors.New("something went wrong with the request"),
}

// BearerTokenError indicates missing or invalid bearer user token.
type BearerTokenError struct {
	Err errors.Error
}

func (ebte *BearerTokenError) Error() string {
	return ebte.Err.Error()
}

var ErrBearerToken = &BearerTokenError{
	Err: errors.New("missing or invalid bearer user token"),
}

// BearerKeyError indicates missing or invalid bearer entity key.
type BearerKeyError struct {
	Err errors.Error
}

func (ebke *BearerKeyError) Error() string {
	return ebke.Err.Error()
}

var ErrBearerKey = &BearerKeyError{
	Err: errors.New("missing or invalid bearer entity key"),
}

// MissingIDError indicates missing entity ID.
type MissingIDError struct {
	Err errors.Error
}

func (mide *MissingIDError) Error() string {
	return mide.Err.Error()
}

var ErrMissingID = &MissingIDError{
	Err: errors.New("missing entity id"),
}

// InvalidAuthKeyError indicates invalid auth key.
type InvalidAuthKeyError struct {
	Err errors.Error
}

func (iahe *InvalidAuthKeyError) Error() string {
	return iahe.Err.Error()
}

var ErrInvalidAuthKey = &InvalidAuthKeyError{
	Err: errors.New("invalid auth key"),
}

// InvalidIDFormatError indicates an invalid ID format.
type InvalidIDFormatError struct {
	Err errors.Error
}

func (iife *InvalidIDFormatError) Error() string {
	return iife.Err.Error()
}

var ErrInvalidIDFormat = &InvalidIDFormatError{
	Err: errors.New("invalid id format provided"),
}

// NameSizeError indicates that name size exceeds the max.
type NameSizeError struct {
	Err errors.Error
}

func (nse *NameSizeError) Error() string {
	return nse.Err.Error()
}

var ErrNameSize = &NameSizeError{
	Err: errors.New("invalid name size"),
}

// EmailSizeError indicates that email size exceeds the max.
type EmailSizeError struct {
	Err errors.Error
}

func (ese *EmailSizeError) Error() string {
	return ese.Err.Error()
}

var ErrEmailSize = &EmailSizeError{
	Err: errors.New("invalid email size"),
}

// InvalidRoleError indicates that an invalid role.
type InvalidRoleError struct {
	Err errors.Error
}

func (ire *InvalidRoleError) Error() string {
	return ire.Err.Error()
}

var ErrInvalidRole = &InvalidRoleError{
	Err: errors.New("invalid client role"),
}

// LimitSizeError indicates that an invalid limit.
type LimitSizeError struct {
	Err errors.Error
}

func (lse *LimitSizeError) Error() string {
	return lse.Err.Error()
}

var ErrLimitSize = &LimitSizeError{
	Err: errors.New("invalid limit size"),
}

// OffsetSizeError indicates an invalid offset.
type OffsetSizeError struct {
	Err errors.Error
}

func (ose *OffsetSizeError) Error() string {
	return ose.Err.Error()
}

var ErrOffsetSize = &OffsetSizeError{
	Err: errors.New("invalid offset size"),
}

// InvalidOrderError indicates an invalid list order.
type InvalidOrderError struct {
	Err errors.Error
}

func (ioe *InvalidOrderError) Error() string {
	return ioe.Err.Error()
}

var ErrInvalidOrder = &InvalidOrderError{
	Err: errors.New("invalid list order provided"),
}

// InvalidDirectionError indicates an invalid list direction.
type InvalidDirectionError struct {
	Err errors.Error
}

func (ide *InvalidDirectionError) Error() string {
	return ide.Err.Error()
}

var ErrInvalidDirection = &InvalidDirectionError{
	Err: errors.New("invalid list direction provided"),
}

// InvalidMemberKindError indicates an invalid member kind.
type InvalidMemberKindError struct {
	Err errors.Error
}

func (imke *InvalidMemberKindError) Error() string {
	return imke.Err.Error()
}

var ErrInvalidMemberKind = &InvalidMemberKindError{
	Err: errors.New("invalid member kind"),
}

// EmptyListError indicates that entity data is empty.
type EmptyListError struct {
	Err errors.Error
}

func (ele *EmptyListError) Error() string {
	return ele.Err.Error()
}

var ErrEmptyList = &EmptyListError{
	Err: errors.New("empty list provided"),
}

// MalformedPolicyError indicates that policies are malformed.
type MalformedPolicyError struct {
	Err errors.Error
}

func (mpe *MalformedPolicyError) Error() string {
	return mpe.Err.Error()
}

var ErrMalformedPolicy = &MalformedPolicyError{
	Err: errors.New("malformed policy"),
}

// MissingPolicySubError indicates that policies are subject.
type MissingPolicySubError struct {
	Err errors.Error
}

func (mpse *MissingPolicySubError) Error() string {
	return mpse.Err.Error()
}

var ErrMissingPolicySub = &MissingPolicySubError{
	Err: errors.New("malformed policy subject"),
}

// MissingPolicyObjError indicates missing policies object.
type MissingPolicyObjError struct {
	Err errors.Error
}

func (mpoe *MissingPolicyObjError) Error() string {
	return mpoe.Err.Error()
}

var ErrMissingPolicyObj = &MissingPolicyObjError{
	Err: errors.New("malformed policy object"),
}

// MalformedPolicyActError indicates missing policies action.
type MalformedPolicyActError struct {
	Err errors.Error
}

func (mpae *MalformedPolicyActError) Error() string {
	return mpae.Err.Error()
}

var ErrMalformedPolicyAct = &MalformedPolicyActError{
	Err: errors.New("malformed policy action"),
}

// MalformedPolicyPerError indicates missing policies relation.
type MalformedPolicyPerError struct {
	Err errors.Error
}

func (mppe *MalformedPolicyPerError) Error() string {
	return mppe.Err.Error()
}

var ErrMalformedPolicyPer = &MalformedPolicyPerError{
	Err: errors.New("malformed policy permission"),
}

// MissingCertDataError indicates missing cert data (ttl).
type MissingCertDataError struct {
	Err errors.Error
}

func (mcde *MissingCertDataError) Error() string {
	return mcde.Err.Error()
}

var ErrMissingCertData = &MissingCertDataError{
	Err: errors.New("missing certificate data"),
}

// InvalidCertDataError indicates invalid cert data (ttl).
type InvalidCertDataError struct {
	Err errors.Error
}

func (icde *InvalidCertDataError) Error() string {
	return icde.Err.Error()
}

var ErrInvalidCertData = &InvalidCertDataError{
	Err: errors.New("invalid certificate data"),
}

// InvalidTopicError indicates an invalid subscription topic.
type InvalidTopicError struct {
	Err errors.Error
}

func (ite *InvalidTopicError) Error() string {
	return ite.Err.Error()
}

var ErrInvalidTopic = &InvalidTopicError{
	Err: errors.New("invalid subscription topic"),
}

// InvalidContactError indicates an invalid subscription contract.
type InvalidContactError struct {
	Err errors.Error
}

func (ice *InvalidContactError) Error() string {
	return ice.Err.Error()
}

var ErrInvalidContact = &InvalidContactError{
	Err: errors.New("invalid subscription contact"),
}

// MissingEmailError indicates missing email.
type MissingEmailError struct {
	Err errors.Error
}

func (mee *MissingEmailError) Error() string {
	return mee.Err.Error()
}

var ErrMissingEmail = &MissingEmailError{
	Err: errors.New("missing email"),
}

// MissingHostError indicates missing host.
type MissingHostError struct {
	Err errors.Error
}

func (mhe *MissingHostError) Error() string {
	return mhe.Err.Error()
}

var ErrMissingHost = &MissingHostError{
	Err: errors.New("missing host"),
}

// MissingPassError indicates missing password.
type MissingPassError struct {
	Err errors.Error
}

func (mpe *MissingPassError) Error() string {
	return mpe.Err.Error()
}

var ErrMissingPass = &MissingPassError{
	Err: errors.New("missing password"),
}

// MissingConfPassError indicates missing conf password.
type MissingConfPassError struct {
	Err errors.Error
}

func (mcpe *MissingConfPassError) Error() string {
	return mcpe.Err.Error()
}

var ErrMissingConfPass = &MissingConfPassError{
	Err: errors.New("missing conf password"),
}

// InvalidResetPassError indicates an invalid reset password.
type InvalidResetPassError struct {
	Err errors.Error
}

func (irpe *InvalidResetPassError) Error() string {
	return irpe.Err.Error()
}

var ErrInvalidResetPass = &InvalidResetPassError{
	Err: errors.New("invalid reset password"),
}

// InvalidComparatorError indicates an invalid comparator.
type InvalidComparatorError struct {
	Err errors.Error
}

func (ice *InvalidComparatorError) Error() string {
	return ice.Err.Error()
}

var ErrInvalidComparator = &InvalidComparatorError{
	Err: errors.New("invalid comparator"),
}

// MissingMemberTypeError indicates missing group member type.
type MissingMemberTypeError struct {
	Err errors.Error
}

func (mmte *MissingMemberTypeError) Error() string {
	return mmte.Err.Error()
}

var ErrMissingMemberType = &MissingMemberTypeError{
	Err: errors.New("missing group member type"),
}

// MissingMemberKindError indicates missing group member kind.
type MissingMemberKindError struct {
	Err errors.Error
}

func (mmke *MissingMemberKindError) Error() string {
	return mmke.Err.Error()
}

var ErrMissingMemberKind = &MissingMemberKindError{
	Err: errors.New("missing group member kind"),
}

// MissingRelationError indicates missing relation.
type MissingRelationError struct {
	Err errors.Error
}

func (mre *MissingRelationError) Error() string {
	return mre.Err.Error()
}

var ErrMissingRelation = &MissingRelationError{
	Err: errors.New("missing relation"),
}

// InvalidRelationError indicates an invalid relation.
type InvalidRelationError struct {
	Err errors.Error
}

func (ire *InvalidRelationError) Error() string {
	return ire.Err.Error()
}

var ErrInvalidRelation = &InvalidRelationError{
	Err: errors.New("invalid relation"),
}

// InvalidAPIKeyError indicates an invalid API key type.
type InvalidAPIKeyError struct {
	Err errors.Error
}

func (iake *InvalidAPIKeyError) Error() string {
	return iake.Err.Error()
}

var ErrInvalidAPIKey = &InvalidAPIKeyError{
	Err: errors.New("invalid api key type"),
}

// BootstrapStateError indicates an invalid bootstrap state.
type BootstrapStateError struct {
	Err errors.Error
}

func (bse *BootstrapStateError) Error() string {
	return bse.Err.Error()
}

var ErrBootstrapState = &BootstrapStateError{
	Err: errors.New("invalid bootstrap state"),
}

// InvitationStateError indicates an invalid invitation state.
type InvitationStateError struct {
	Err errors.Error
}

func (ise *InvitationStateError) Error() string {
	return ise.Err.Error()
}

var ErrInvitationState = &InvitationStateError{
	Err: errors.New("invalid invitation state"),
}

// MissingIdentityError indicates missing entity Identity.
type MissingIdentityError struct {
	Err errors.Error
}

func (mie *MissingIdentityError) Error() string {
	return mie.Err.Error()
}

var ErrMissingIdentity = &MissingIdentityError{
	Err: errors.New("missing entity identity"),
}

// MissingSecretError indicates missing secret.
type MissingSecretError struct {
	Err errors.Error
}

func (mse *MissingSecretError) Error() string {
	return mse.Err.Error()
}

var ErrMissingSecret = &MissingSecretError{
	Err: errors.New("missing secret"),
}

// PasswordFormatError indicates weak password.
type PasswordFormatError struct {
	Err errors.Error
}

func (pfe *PasswordFormatError) Error() string {
	return pfe.Err.Error()
}

var ErrPasswordFormat = &PasswordFormatError{
	Err: errors.New("password does not meet the requirements"),
}

// MissingNameError indicates missing identity name.
type MissingNameError struct {
	Err errors.Error
}

func (mne *MissingNameError) Error() string {
	return mne.Err.Error()
}

var ErrMissingName = &MissingNameError{
	Err: errors.New("missing identity name"),
}

// InvalidLevelError indicates an invalid group level.
type InvalidLevelError struct {
	Err errors.Error
}

func (ile *InvalidLevelError) Error() string {
	return ile.Err.Error()
}

var ErrInvalidLevel = &InvalidLevelError{
	Err: errors.New("invalid group level (should be between 0 and 5)"),
}

// NotFoundParamError indicates that the parameter was not found in the query.
type NotFoundParamError struct {
	Err errors.Error
}

func (nfpe *NotFoundParamError) Error() string {
	return nfpe.Err.Error()
}

var ErrNotFoundParam = &NotFoundParamError{
	Err: errors.New("parameter not found in the query"),
}

// InvalidQueryParamsError indicates invalid query parameters.
type InvalidQueryParamsError struct {
	Err errors.Error
}

func (iqpe *InvalidQueryParamsError) Error() string {
	return iqpe.Err.Error()
}

var ErrInvalidQueryParams = &InvalidQueryParamsError{
	Err: errors.New("invalid query parameters"),
}

// InvalidVisibilityTypeError indicates invalid visibility type.
type InvalidVisibilityTypeError struct {
	Err errors.Error
}

func (ivte *InvalidVisibilityTypeError) Error() string {
	return ivte.Err.Error()
}

var ErrInvalidVisibilityType = &InvalidVisibilityTypeError{
	Err: errors.New("invalid visibility type"),
}

// UnsupportedContentTypeError indicates unacceptable or lack of Content-Type.
type UnsupportedContentTypeError struct {
	Err errors.Error
}

func (ucte *UnsupportedContentTypeError) Error() string {
	return ucte.Err.Error()
}

var ErrUnsupportedContentType = &UnsupportedContentTypeError{
	Err: errors.New("unsupported content type"),
}

// RollbackTxError indicates failed to rollback transaction.
type RollbackTxError struct {
	Err errors.Error
}

func (rte *RollbackTxError) Error() string {
	return rte.Err.Error()
}

var ErrRollbackTx = &RollbackTxError{
	Err: errors.New("failed to rollback transaction"),
}

// InvalidAggregationError indicates invalid aggregation value.
type InvalidAggregationError struct {
	Err errors.Error
}

func (iae *InvalidAggregationError) Error() string {
	return iae.Err.Error()
}

var ErrInvalidAggregation = &InvalidAggregationError{
	Err: errors.New("invalid aggregation value"),
}

// InvalidIntervalError indicates invalid interval value.
type InvalidIntervalError struct {
	Err errors.Error
}

func (iie *InvalidIntervalError) Error() string {
	return iie.Err.Error()
}

var ErrInvalidInterval = &InvalidIntervalError{
	Err: errors.New("invalid interval value"),
}

// MissingFromError indicates missing from value.
type MissingFromError struct {
	Err errors.Error
}

func (mfe *MissingFromError) Error() string {
	return mfe.Err.Error()
}

var ErrMissingFrom = &MissingFromError{
	Err: errors.New("missing from time value"),
}

// MissingToError indicates missing to value.
type MissingToError struct {
	Err errors.Error
}

func (mte *MissingToError) Error() string {
	return mte.Err.Error()
}

var ErrMissingTo = &MissingToError{
	Err: errors.New("missing to time value"),
}

// EmptyMessageError indicates empty message.
type EmptyMessageError struct {
	Err errors.Error
}

func (eme *EmptyMessageError) Error() string {
	return eme.Err.Error()
}

var ErrEmptyMessage = &EmptyMessageError{
	Err: errors.New("empty message"),
}
