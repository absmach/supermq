package authn

import (
	"context"
	"net/http"
	"os"
	"strconv"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/auth"
	"github.com/go-chi/chi/v5"
)

type sessionKeyType string

const (
	allowUnverifiedUserEnv = "SMQ_ALLOW_UNVERIFIED_USER"

	SessionKey = sessionKeyType("session")
)

// middlewareOptions contains configuration for authentication middleware.
type middlewareOptions struct {
	domainCheck         bool
	allowUnverifiedUser bool
}

// defaultMiddlewareOptions returns the default middleware configuration.
func defaultMiddlewareOptions() *middlewareOptions {
	return &middlewareOptions{
		domainCheck:         true,
		allowUnverifiedUser: false,
	}
}

// MiddlewareOption is a function that modifies middleware options.
type MiddlewareOption func(*middlewareOptions)

// WithDomainCheck sets whether domain checking is enabled.
func WithDomainCheck(enabled bool) MiddlewareOption {
	return func(opts *middlewareOptions) {
		opts.domainCheck = enabled
	}
}

// WithAllowUnverifiedUser sets whether unverified users are allowed.
func WithAllowUnverifiedUser(allowed bool) MiddlewareOption {
	return func(opts *middlewareOptions) {
		opts.allowUnverifiedUser = allowed
	}
}

// WithDefaultMiddlewareOptions resets options to default values.
func WithDefaultMiddlewareOptions() MiddlewareOption {
	return func(opts *middlewareOptions) {
		defaults := defaultMiddlewareOptions()
		opts.domainCheck = defaults.domainCheck
		opts.allowUnverifiedUser = defaults.allowUnverifiedUser
	}
}

// Authn defines the interface for authenticated services with middleware.
type Authn interface {
	Authentication
	WithOptions(options ...MiddlewareOption) Authn
	Middleware() func(http.Handler) http.Handler
}

// authnService wraps Authentication with middleware functionality.
type authnService struct {
	Authentication
	options []MiddlewareOption
}

// NewAuthn creates a new authenticated service with middleware support.
// The order of precedence for options is as follows, with later options overriding earlier ones:
// 1. Default options (lowest precedence).
// 2. Options from environment variables (e.g., SMQ_ALLOW_UNVERIFIED_USER).
// 3. Options passed as arguments to this function (highest precedence).
//
// For example, consider the 'allowUnverifiedUser' option:
//   - By default, it is 'false'.
//   - If the SMQ_ALLOW_UNVERIFIED_USER environment variable is set to "true",
//     it becomes 'true'.
//   - If NewAuthn is called with WithAllowUnverifiedUser(false), it will be 'false',
//     regardless of the environment variable, as function arguments have the highest precedence.
func NewAuthn(authn Authentication, options ...MiddlewareOption) Authn {
	allOptions := []MiddlewareOption{WithDefaultMiddlewareOptions()}
	if val, ok := os.LookupEnv(allowUnverifiedUserEnv); ok {
		allowUnverifiedUser, err := strconv.ParseBool(val)
		if err == nil && allowUnverifiedUser {
			allOptions = append(allOptions, WithAllowUnverifiedUser(true))
		}
	}
	allOptions = append(allOptions, options...)
	return &authnService{
		Authentication: authn,
		options:        allOptions,
	}
}

// WithOptions returns a new service with additional options.
func (a *authnService) WithOptions(options ...MiddlewareOption) Authn {
	return &authnService{
		Authentication: a.Authentication,
		options:        append(a.options, options...),
	}
}

// getMiddlewareOptions returns the configured middleware options.
func (a *authnService) getMiddlewareOptions() *middlewareOptions {
	opts := defaultMiddlewareOptions()
	for _, option := range a.options {
		option(opts)
	}
	return opts
}

// Middleware returns an HTTP middleware function that handles authentication.
func (a *authnService) Middleware() func(http.Handler) http.Handler {
	opts := a.getMiddlewareOptions()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := apiutil.ExtractBearerToken(r)
			if token == "" {
				http.Error(w, apiutil.ErrBearerToken.Error(), http.StatusUnauthorized)
				return
			}

			resp, err := a.Authenticate(r.Context(), token)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if !opts.allowUnverifiedUser && !resp.Verified {
				http.Error(w, "email not verified", http.StatusUnauthorized)
				return
			}

			if opts.domainCheck {
				domain := chi.URLParam(r, "domainID")
				if domain == "" {
					http.Error(w, "missing domain ID", http.StatusBadRequest)
					return
				}
				resp.DomainID = domain
				switch resp.Role {
				case AdminRole:
					resp.DomainUserID = resp.UserID
				case UserRole:
					resp.DomainUserID = auth.EncodeDomainUserID(domain, resp.UserID)
				}
			}

			ctx := context.WithValue(r.Context(), SessionKey, resp)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
