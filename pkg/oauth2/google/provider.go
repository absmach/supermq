// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package google

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	mgoauth2 "github.com/absmach/supermq/pkg/oauth2"
	uclient "github.com/absmach/supermq/users"
	"golang.org/x/oauth2"
	googleoauth2 "golang.org/x/oauth2/google"
)

const (
	providerName = "google"
	defTimeout   = 1 * time.Minute
	userInfoURL  = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	tokenInfoURL = "https://oauth2.googleapis.com/tokeninfo?access_token="
)

var scopes = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

var httpClient = &http.Client{
	Timeout: defTimeout,
}

var _ mgoauth2.Provider = (*provider)(nil)

type provider struct {
	config        *oauth2.Config
	state         string
	uiRedirectURL string
	errorURL      string
}

// NewProvider returns a new Google OAuth provider.
func NewProvider(cfg mgoauth2.Config, uiRedirectURL, errorURL string) mgoauth2.Provider {
	return &provider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint:     googleoauth2.Endpoint,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       scopes,
		},
		state:         cfg.State,
		uiRedirectURL: uiRedirectURL,
		errorURL:      errorURL,
	}
}

func (cfg *provider) Name() string {
	return providerName
}

func (cfg *provider) State() string {
	return cfg.state
}

func (cfg *provider) RedirectURL() string {
	return cfg.uiRedirectURL
}

func (cfg *provider) ErrorURL() string {
	return cfg.errorURL
}

func (cfg *provider) IsEnabled() bool {
	return cfg.config.ClientID != "" && cfg.config.ClientSecret != ""
}

func (cfg *provider) Exchange(ctx context.Context, code string) (oauth2.Token, error) {
	token, err := cfg.config.Exchange(ctx, code)
	if err != nil {
		return oauth2.Token{}, err
	}

	return *token, nil
}

func (cfg *provider) ExchangeWithRedirect(ctx context.Context, code, redirectURL string) (oauth2.Token, error) {
	// Create a temporary config with the custom redirect URL
	tempConfig := *cfg.config
	tempConfig.RedirectURL = redirectURL
	token, err := tempConfig.Exchange(ctx, code)
	if err != nil {
		return oauth2.Token{}, err
	}

	return *token, nil
}

func (cfg *provider) UserInfo(accessToken string) (uclient.User, error) {
	resp, err := httpClient.Get(userInfoURL + url.QueryEscape(accessToken))
	if err != nil {
		return uclient.User{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uclient.User{}, svcerr.ErrAuthentication
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return uclient.User{}, err
	}

	user, err := mgoauth2.NormalizeUser(data, providerName)
	if err != nil {
		return uclient.User{}, errors.Wrap(err, svcerr.ErrAuthentication)
	}

	return user, nil
}

func (cfg *provider) GetAuthURL() string {
	return cfg.config.AuthCodeURL(cfg.state)
}

func (cfg *provider) GetAuthURLWithRedirect(redirectURL string) string {
	// Create a temporary config with the custom redirect URL
	tempConfig := *cfg.config
	tempConfig.RedirectURL = redirectURL
	return tempConfig.AuthCodeURL(cfg.state)
}
