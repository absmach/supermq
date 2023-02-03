package jwt

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/mainflux/mainflux/pkg/errors"
)

const issuerName = "clients.auth"

var _ TokenRepository = (*tokenRepo)(nil)

var (
	accessDuration  time.Duration = time.Hour * 15
	refreshDuration time.Duration = time.Hour * 24
)

type tokenRepo struct {
	secret []byte
}

// NewTokenRepo instantiates an implementation of Token repository.
func NewTokenRepo(secret []byte) TokenRepository {
	return &tokenRepo{
		secret: secret,
	}
}

func (repo tokenRepo) Issue(ctx context.Context, claim Claims) (Token, error) {
	aexpiry := time.Now().Add(accessDuration)
	accessToken, err := jwt.NewBuilder().
		Issuer(issuerName).
		IssuedAt(time.Now()).
		Subject(claim.ClientID).
		Claim("type", AccessToken).
		Claim("role", claim.Role).
		Claim("tag", claim.Tag).
		Expiration(aexpiry).
		Build()
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	signedAccessToken, err := jwt.Sign(accessToken, jwt.WithKey(jwa.HS512, repo.secret))
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	refreshToken, err := jwt.NewBuilder().
		Issuer(issuerName).
		IssuedAt(time.Now()).
		Subject(claim.ClientID).
		Claim("type", RefreshToken).
		Claim("role", claim.Role).
		Claim("tag", claim.Tag).
		Expiration(time.Now().Add(refreshDuration)).
		Build()
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	signedRefreshToken, err := jwt.Sign(refreshToken, jwt.WithKey(jwa.HS512, repo.secret))
	if err != nil {
		return Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	return Token{
		AccessToken:  string(signedAccessToken[:]),
		RefreshToken: string(signedRefreshToken[:]),
		AccessType:   "Bearer",
	}, nil
}

func (repo tokenRepo) Parse(ctx context.Context, accessToken string) (Claims, error) {
	token, err := jwt.Parse(
		[]byte(accessToken),
		jwt.WithValidate(true),
		jwt.WithKey(jwa.HS512, repo.secret),
	)
	if err != nil {
		return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	tType, ok := token.Get("type")
	if !ok {
		return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	role, ok := token.Get("role")
	if !ok {
		return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	tag, ok := token.Get("tag")
	if !ok {
		return Claims{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	claim := Claims{
		ClientID: token.Subject(),
		Role:     role.(string),
		Tag:      tag.(string),
		Type:     tType.(string),
	}
	return claim, nil
}
