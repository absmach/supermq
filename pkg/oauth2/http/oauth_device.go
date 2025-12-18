// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	oauth2 "github.com/absmach/supermq/pkg/oauth2"
	"github.com/go-chi/chi/v5"
)

var (
	errDeviceCodeExpired = newErrorResponse("device code expired")
	errDeviceCodePending = newErrorResponse("authorization pending")
	errSlowDown          = newErrorResponse("slow down")
	errAccessDenied      = newErrorResponse("access denied")
)

// DeviceHandler registers device flow routes for OAuth2 providers.
func DeviceHandler(r *chi.Mux, tokenClient grpcTokenV1.TokenServiceClient, oauthSvc oauth2.Service, providers ...oauth2.Provider) *chi.Mux {
	for _, provider := range providers {
		r.Post("/oauth/device/code/"+provider.Name(), DeviceCodeHandler(provider, oauthSvc))
		r.Post("/oauth/device/token/"+provider.Name(), DeviceTokenHandler(provider, oauthSvc))
	}
	// Register verify endpoints once (not per provider)
	r.Get("/oauth/device/verify", DeviceVerifyPageHandler())
	r.Post("/oauth/device/verify", DeviceVerifyHandler(oauthSvc, providers...))
	return r
}

// DeviceCodeHandler initiates the device authorization flow.
func DeviceCodeHandler(provider oauth2.Provider, oauthSvc oauth2.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !provider.IsEnabled() {
			errResp := errProviderDisabled
			respondWithJSON(w, http.StatusNotFound, errResp)
			return
		}

		// Build verification URI with proper scheme
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		verificationURI := fmt.Sprintf("%s://%s/oauth/device/verify", scheme, r.Host)

		code, err := oauthSvc.CreateDeviceCode(r.Context(), provider, verificationURI)
		if err != nil {
			respondWithJSON(w, http.StatusInternalServerError, newErrorResponse(err.Error()))
			return
		}

		respondWithJSON(w, http.StatusOK, code)
	}
}

// DeviceTokenHandler polls for device authorization completion.
func DeviceTokenHandler(provider oauth2.Provider, oauthSvc oauth2.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !provider.IsEnabled() {
			respondWithJSON(w, http.StatusNotFound, errProviderDisabled)
			return
		}

		var req struct {
			DeviceCode string `json:"device_code"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithJSON(w, http.StatusBadRequest, errInvalidBody)
			return
		}

		jwt, err := oauthSvc.PollDeviceToken(r.Context(), provider, req.DeviceCode)
		if err != nil {
			// Map OAuth service errors to appropriate HTTP responses
			switch {
			case errors.Is(err, oauth2.ErrDeviceCodeNotFound):
				respondWithJSON(w, http.StatusNotFound, newErrorResponse("invalid device code"))
			case errors.Is(err, oauth2.ErrDeviceCodeExpired):
				respondWithJSON(w, http.StatusBadRequest, errDeviceCodeExpired)
			case errors.Is(err, oauth2.ErrSlowDown):
				respondWithJSON(w, http.StatusBadRequest, errSlowDown)
			case errors.Is(err, oauth2.ErrAccessDenied):
				respondWithJSON(w, http.StatusUnauthorized, errAccessDenied)
			case errors.Is(err, oauth2.ErrDeviceCodePending):
				respondWithJSON(w, http.StatusAccepted, errDeviceCodePending)
			default:
				respondWithJSON(w, http.StatusInternalServerError, newErrorResponse(err.Error()))
			}
			return
		}

		jwt.AccessType = ""
		respondWithJSON(w, http.StatusOK, jwt)
	}
}

// DeviceVerifyPageHandler serves the HTML page for device verification.
func DeviceVerifyPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, deviceVerifyHTML)
	}
}

// DeviceVerifyHandler handles user verification of device codes.
func DeviceVerifyHandler(oauthSvc oauth2.Service, providers ...oauth2.Provider) http.HandlerFunc {
	providerMap := make(map[string]oauth2.Provider)
	for _, p := range providers {
		providerMap[p.Name()] = p
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req struct {
			UserCode string `json:"user_code"`
			Code     string `json:"code"`
			Approve  bool   `json:"approve"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithJSON(w, http.StatusBadRequest, errInvalidBody)
			return
		}

		code, err := oauthSvc.GetDeviceCodeByUserCode(r.Context(), req.UserCode)
		if err != nil {
			if errors.Is(err, oauth2.ErrUserCodeNotFound) {
				respondWithJSON(w, http.StatusNotFound, newErrorResponse("invalid user code"))
			} else {
				respondWithJSON(w, http.StatusInternalServerError, newErrorResponse(err.Error()))
			}
			return
		}

		provider, ok := providerMap[code.Provider]
		if !ok {
			respondWithJSON(w, http.StatusBadRequest, newErrorResponse("invalid provider"))
			return
		}

		if !provider.IsEnabled() {
			respondWithJSON(w, http.StatusNotFound, errProviderDisabled)
			return
		}

		if !req.Approve {
			// User denied - pass empty code and approve=false
			if err := oauthSvc.VerifyDevice(r.Context(), provider, req.UserCode, "", false); err != nil {
				respondWithJSON(w, http.StatusInternalServerError, newErrorResponse(err.Error()))
				return
			}
			respondWithJSON(w, http.StatusOK, map[string]string{"status": "denied"})
			return
		}

		// User approved - verify with the OAuth code
		if err := oauthSvc.VerifyDevice(r.Context(), provider, req.UserCode, req.Code, true); err != nil {
			if errors.Is(err, oauth2.ErrDeviceCodeExpired) {
				respondWithJSON(w, http.StatusBadRequest, errDeviceCodeExpired)
			} else {
				respondWithJSON(w, http.StatusUnauthorized, newErrorResponse(err.Error()))
			}
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"status": "approved"})
	}
}
