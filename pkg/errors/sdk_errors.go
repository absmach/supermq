// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"net/http"
)

var (
	errJSONKeyNotFound = New("response body expected error message json key not found")
)

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

func (ce *sdkError) Error() string {
	if ce == nil {
		return ""
	}
	if ce.err == nil {
		return ce.msg
	}
	return ce.msg + " : " + ce.err.Error()
}

func (ce *sdkError) Msg() string {
	return ce.msg
}

func (ce *sdkError) Err() Error {
	return ce.err
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

// CheckError will check for error in http response.
func CheckError(resp *http.Response, expectedStatusCodes ...int) SDKError {
	for _, expectedStatusCode := range expectedStatusCodes {
		if resp.StatusCode == expectedStatusCode {
			return nil
		}
	}

	var content map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return NewSDKErrorWithStatus(err.Error(), resp.StatusCode)
	}

	if msg, ok := content["error"]; ok {
		if v, ok := msg.(string); ok {
			return NewSDKErrorWithStatus(v, resp.StatusCode)
		}
		return NewSDKErrorWithStatus("unknown error", resp.StatusCode)
	}
	return NewSDKErrorWithStatus(errJSONKeyNotFound.Error(), resp.StatusCode)

}
