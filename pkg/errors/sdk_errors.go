// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"io"
	"net/http"
)

var (
	errFailedToReadBody = New("failed to read http response body")
	errRespBodyNotJSON  = New("response body is not a valid JSON")
	errJSONKeyNotFound  = New("response body expected error message json key not found")
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
func CheckError(resp *http.Response, expectedStatusCodes ...int) error {
	for _, expectedStatusCode := range expectedStatusCodes {
		if resp.StatusCode == expectedStatusCode {
			return nil
		}
	}

	b, bErr := io.ReadAll(resp.Body)
	if bErr != nil {
		e := Wrap(errFailedToReadBody, bErr)
		return Wrap(NewSDKErrorWithStatus("", resp.StatusCode), e)
	}

	var content map[string]interface{}
	err := json.Unmarshal(b, &content)
	if err != nil {
		e := Wrap(errRespBodyNotJSON, New(string(b)))
		return Wrap(NewSDKErrorWithStatus("", resp.StatusCode), e)
	}

	if msg, ok := content["error"]; ok {
		if v, ok := msg.(string); ok {
			return NewSDKErrorWithStatus(v, resp.StatusCode)
		}
		return NewSDKErrorWithStatus("unknown error", resp.StatusCode)
	}
	e := Wrap(errJSONKeyNotFound, New(string(b)))
	return Wrap(NewSDKErrorWithStatus("", resp.StatusCode), e)
}
