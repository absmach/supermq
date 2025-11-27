// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/absmach/supermq/users"
	useroauth "github.com/absmach/supermq/users/oauth"
	"github.com/go-chi/chi/v5"
)

var (
	errDeviceCodeExpired = newErrorResponse("device code expired")
	errDeviceCodePending = newErrorResponse("authorization pending")
	errSlowDown          = newErrorResponse("slow down")
	errAccessDenied      = newErrorResponse("access denied")
)

// oauthDeviceHandler registers device flow routes for OAuth2 providers.
func oauthDeviceHandler(r *chi.Mux, svc users.Service, tokenClient grpcTokenV1.TokenServiceClient, oauthSvc useroauth.Service, providers ...oauth2.Provider) *chi.Mux {
	for _, provider := range providers {
		r.Post("/oauth/device/code/"+provider.Name(), deviceCodeHandler(provider, oauthSvc))
		r.Post("/oauth/device/token/"+provider.Name(), deviceTokenHandler(provider, oauthSvc))
	}
	// Register verify endpoints once (not per provider)
	r.Get("/oauth/device/verify", deviceVerifyPageHandler())
	r.Post("/oauth/device/verify", deviceVerifyHandler(oauthSvc, providers...))
	return r
}

// deviceCodeHandler initiates the device authorization flow.
func deviceCodeHandler(provider oauth2.Provider, oauthSvc useroauth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !provider.IsEnabled() {
			errResp := newErrorResponse("oauth provider is disabled")
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

// deviceTokenHandler polls for device authorization completion.
func deviceTokenHandler(provider oauth2.Provider, oauthSvc useroauth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if !provider.IsEnabled() {
			errResp := newErrorResponse("oauth provider is disabled")
			respondWithJSON(w, http.StatusNotFound, errResp)
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
			case errors.Is(err, useroauth.ErrDeviceCodeNotFound):
				respondWithJSON(w, http.StatusNotFound, newErrorResponse("invalid device code"))
			case errors.Is(err, useroauth.ErrDeviceCodeExpired):
				respondWithJSON(w, http.StatusBadRequest, errDeviceCodeExpired)
			case errors.Is(err, useroauth.ErrSlowDown):
				respondWithJSON(w, http.StatusBadRequest, errSlowDown)
			case errors.Is(err, useroauth.ErrAccessDenied):
				respondWithJSON(w, http.StatusUnauthorized, errAccessDenied)
			case errors.Is(err, useroauth.ErrDeviceCodePending):
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

// deviceVerifyPageHandler serves the HTML page for device verification.
func deviceVerifyPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Device Verification - Magistrala</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #073764;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 48px;
            max-width: 500px;
            width: 100%;
            text-align: center;
        }
        h1 {
            color: #073764;
            font-size: 32px;
            margin-bottom: 24px;
        }
        p {
            color: #4a5568;
            font-size: 18px;
            line-height: 1.6;
            margin-bottom: 24px;
        }
        .code-input {
            width: 100%;
            padding: 16px;
            font-size: 24px;
            text-align: center;
            border: 2px solid #e2e8f0;
            border-radius: 8px;
            margin-bottom: 24px;
            text-transform: uppercase;
            letter-spacing: 4px;
        }
        .code-input:focus {
            outline: none;
            border-color: #073764;
        }
        button {
            width: 100%;
            padding: 16px;
            font-size: 18px;
            font-weight: 600;
            color: white;
            background: #073764;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            transition: background 0.2s;
        }
        button:hover {
            background: #05294a;
        }
        button:disabled {
            background: #cbd5e0;
            cursor: not-allowed;
        }
        .message {
            margin-top: 24px;
            padding: 16px;
            border-radius: 8px;
            display: none;
        }
        .message.success {
            background: #d1fae5;
            color: #065f46;
            display: block;
        }
        .message.error {
            background: #fee2e2;
            color: #991b1b;
            display: block;
        }
        @media (max-width: 600px) {
            .container { padding: 32px 24px; }
            h1 { font-size: 28px; }
            .code-input { font-size: 20px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Device Verification</h1>
        <p>Enter the code displayed on your device</p>
        <form id="verifyForm">
            <input
                type="text"
                id="userCode"
                class="code-input"
                placeholder="XXXX-XXXX"
                maxlength="9"
                pattern="[A-Za-z0-9]{4}-[A-Za-z0-9]{4}"
                required
            />
            <button type="submit" id="submitBtn">Verify Device</button>
        </form>
        <div id="message" class="message"></div>
    </div>

    <script>
        const form = document.getElementById('verifyForm');
        const input = document.getElementById('userCode');
        const submitBtn = document.getElementById('submitBtn');
        const message = document.getElementById('message');

        // Auto-format input
        input.addEventListener('input', (e) => {
            let value = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, '');
            if (value.length > 4) {
                value = value.slice(0, 4) + '-' + value.slice(4, 8);
            }
            e.target.value = value;
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            const userCode = input.value.trim();
            if (!userCode || userCode.length < 9) {
                showMessage('Please enter a valid code', 'error');
                return;
            }

            submitBtn.disabled = true;
            submitBtn.textContent = 'Verifying...';

            try {
                // Get authorization URL for the provider (Google as default)
                // Build the callback URL dynamically from current location
                const callbackURL = window.location.protocol + '//' + window.location.host + '/oauth/callback/google';
                const encodedCallback = encodeURIComponent(callbackURL);

                const authResponse = await fetch('/oauth/authorize/google?redirect_uri=' + encodedCallback);
                if (!authResponse.ok) throw new Error('Failed to get authorization URL');

                const authData = await authResponse.json();

                // Build OAuth URL with device state
                const deviceState = 'device:' + userCode;
                const authURL = authData.authorization_url.replace(/state=[^&]+/, 'state=' + encodeURIComponent(deviceState));

                // Redirect to OAuth provider
                window.location.href = authURL;

            } catch (error) {
                showMessage('Verification failed: ' + error.message, 'error');
                submitBtn.disabled = false;
                submitBtn.textContent = 'Verify Device';
            }
        });

        function showMessage(text, type) {
            message.textContent = text;
            message.className = 'message ' + type;
        }
    </script>
</body>
</html>`
		fmt.Fprint(w, html)
	}
}

// deviceVerifyHandler handles user verification of device codes.
func deviceVerifyHandler(oauthSvc useroauth.Service, providers ...oauth2.Provider) http.HandlerFunc {
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
			if errors.Is(err, useroauth.ErrUserCodeNotFound) {
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
			respondWithJSON(w, http.StatusNotFound, newErrorResponse("oauth provider is disabled"))
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
			if errors.Is(err, useroauth.ErrDeviceCodeExpired) {
				respondWithJSON(w, http.StatusBadRequest, errDeviceCodeExpired)
			} else {
				respondWithJSON(w, http.StatusUnauthorized, newErrorResponse(err.Error()))
			}
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"status": "approved"})
	}
}
