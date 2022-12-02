// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

// SDKError is an error type for Mainflux SDK.
type SDKError interface {
	Error
	StatusCode() int
}

var _ SDKError = (*sdkError)(nil)

type sdkError struct {
	*customError
	statusCode int
}

func (ce *sdkError) StatusCode() int {
	return ce.statusCode
}

// NewSDKError returns an SDK Error that formats as the given text.
func NewSDKError(text string) SDKError {
	return &sdkError{
		customError: &customError{
			msg: text,
			err: nil,
		},
	}
}

// NewSDKErrorWithStatus returns an SDK Error setting the status code.
func NewSDKErrorWithStatus(msg string, statusCode int) SDKError {
	return &sdkError{
		statusCode: statusCode,
		customError: &customError{
			msg: msg,
			err: nil,
		},
	}
}
