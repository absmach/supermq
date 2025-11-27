// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/absmach/supermq/users"
	useroauth "github.com/absmach/supermq/users/oauth"
	"github.com/go-chi/chi/v5"
)

var (
	errInvalidBody  = newErrorResponse("invalid request body")
	errInvalidState = newErrorResponse("invalid state")
	errEmptyCode    = newErrorResponse("empty code")
)

type errorResponse struct {
	Error string `json:"error"`
}

// newErrorResponse creates a JSON error response.
func newErrorResponse(msg string) errorResponse {
	return errorResponse{Error: msg}
}

type authURLResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
}

// oauthHandler registers OAuth2 routes for the given providers.
// It sets up three endpoints for each provider:
// - GET /oauth/authorize/{provider} - Returns the authorization URL
// - GET /oauth/callback/{provider} - Handles OAuth2 callback and sets cookies
// - POST /oauth/cli/callback/{provider} - Handles CLI OAuth2 callback and returns JSON.
func oauthHandler(r *chi.Mux, svc users.Service, tokenClient grpcTokenV1.TokenServiceClient, oauthSvc useroauth.Service, providers ...oauth2.Provider) *chi.Mux {
	for _, provider := range providers {
		r.HandleFunc("/oauth/callback/"+provider.Name(), oauth2CallbackHandler(provider, oauthSvc))
		r.Get("/oauth/authorize/"+provider.Name(), oauth2AuthorizeHandler(provider))
		r.Post("/oauth/cli/callback/"+provider.Name(), oauth2CLICallbackHandler(provider, oauthSvc))
	}

	return r
}

// oauth2CallbackHandler is a http.HandlerFunc that handles OAuth2 callbacks.
func oauth2CallbackHandler(oauth oauth2.Provider, oauthSvc useroauth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !oauth.IsEnabled() {
			redirectWithError(w, r, oauth.ErrorURL(), "oauth provider is disabled")
			return
		}

		state := r.FormValue("state")

		// Check if this is a device flow callback (state contains device: prefix)
		if useroauth.IsDeviceFlowState(state) {
			handleDeviceFlowCallback(w, r, oauth, oauthSvc)
			return
		}

		if state != oauth.State() {
			redirectWithError(w, r, oauth.ErrorURL(), "invalid state")
			return
		}

		code := r.FormValue("code")
		if code == "" {
			redirectWithError(w, r, oauth.ErrorURL(), "empty code")
			return
		}

		jwt, err := oauthSvc.ProcessWebCallback(r.Context(), oauth, code, "")
		if err != nil {
			redirectWithError(w, r, oauth.ErrorURL(), err.Error())
			return
		}

		setTokenCookies(w, jwt)
		http.Redirect(w, r, oauth.RedirectURL(), http.StatusFound)
	}
}

// oauth2AuthorizeHandler returns the authorization URL for the OAuth2 provider.
func oauth2AuthorizeHandler(oauth oauth2.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !oauth.IsEnabled() {
			errResp := newErrorResponse("oauth provider is disabled")
			respondWithJSON(w, http.StatusNotFound, errResp)
			return
		}

		redirectURL := r.URL.Query().Get("redirect_uri")
		var authURL string
		if redirectURL != "" {
			authURL = oauth.GetAuthURLWithRedirect(redirectURL)
		} else {
			authURL = oauth.GetAuthURL()
		}

		resp := authURLResponse{
			AuthorizationURL: authURL,
			State:            oauth.State(),
		}
		respondWithJSON(w, http.StatusOK, resp)
	}
}

// oauth2CLICallbackHandler handles OAuth2 callbacks for CLI and returns JSON tokens.
func oauth2CLICallbackHandler(oauth oauth2.Provider, oauthSvc useroauth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !oauth.IsEnabled() {
			errResp := newErrorResponse("oauth provider is disabled")
			respondWithJSON(w, http.StatusNotFound, errResp)
			return
		}
		var req struct {
			Code        string `json:"code"`
			State       string `json:"state"`
			RedirectURL string `json:"redirect_url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithJSON(w, http.StatusBadRequest, errInvalidBody)
			return
		}

		if req.State != oauth.State() {
			respondWithJSON(w, http.StatusBadRequest, errInvalidState)
			return
		}

		if req.Code == "" {
			respondWithJSON(w, http.StatusBadRequest, errEmptyCode)
			return
		}

		jwt, err := oauthSvc.ProcessWebCallback(r.Context(), oauth, req.Code, req.RedirectURL)
		if err != nil {
			status := http.StatusInternalServerError
			if err.Error() == "unauthorized" {
				status = http.StatusUnauthorized
			}
			respondWithJSON(w, status, newErrorResponse(err.Error()))
			return
		}

		jwt.AccessType = ""
		respondWithJSON(w, http.StatusOK, jwt)
	}
}

// respondWithJSON writes a JSON response with the given status code and data.
func respondWithJSON(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// redirectWithError redirects to the baseURL with an error query parameter.
func redirectWithError(w http.ResponseWriter, r *http.Request, baseURL, errMsg string) {
	redirectURL := fmt.Sprintf("%s?error=%s", baseURL, errMsg)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// setTokenCookies sets the access_token and refresh_token cookies in the response.
func setTokenCookies(w http.ResponseWriter, jwt *grpcTokenV1.Token) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    jwt.GetAccessToken(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    jwt.GetRefreshToken(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})
}

// handleDeviceFlowCallback processes OAuth callback for device authorization flow.
func handleDeviceFlowCallback(w http.ResponseWriter, r *http.Request, oauth oauth2.Provider, oauthSvc useroauth.Service) {
	// Extract user code from state (format: "device:ABCD-EFGH")
	state := r.FormValue("state")
	userCode := useroauth.ExtractUserCodeFromState(state)

	// Get device code by user code to validate it exists
	_, err := oauthSvc.GetDeviceCodeByUserCode(r.Context(), userCode)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, strings.Replace(errorHTML, "{{ERROR_MESSAGE}}", "The device code is invalid or has expired.", 1))
		return
	}

	// Get OAuth authorization code
	code := r.FormValue("code")
	if code == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, strings.Replace(errorHTML, "{{ERROR_MESSAGE}}", "No authorization code received.", 1))
		return
	}

	// Process the device callback
	if err := oauthSvc.ProcessDeviceCallback(r.Context(), oauth, userCode, code); err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, strings.Replace(errorHTML, "{{ERROR_MESSAGE}}", fmt.Sprintf("Failed to process callback: %s.", err.Error()), 1))
		return
	}

	// Show success page
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, successHTML)
}
