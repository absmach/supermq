package api

import (
	"net/http"

	"github.com/mainflux/mainflux/homing"
)

type saveTelemetryRes struct {
	created bool
}

func (res saveTelemetryRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res saveTelemetryRes) Headers() map[string]string {
	if res.created {
		return map[string]string{}
	}

	return map[string]string{}
}

func (res saveTelemetryRes) Empty() bool {
	return true
}

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

type telemetryPageRes struct {
	pageRes
	Telemetry []homing.Telemetry `json:"telemetry"`
}

func (res telemetryPageRes) Code() int {
	return http.StatusOK
}

func (res telemetryPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res telemetryPageRes) Empty() bool {
	return false
}
