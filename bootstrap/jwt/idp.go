package jwt

import (
	"errors"
	"fmt"
	"nov/bootstrap"

	jwt "github.com/dgrijalva/jwt-go"
)

var errInvalidJWTClaims = errors.New("error validating claims type")
var _ bootstrap.IdentityProvider = (*jwtIDP)(nil)

type jwtIDP struct{}

// New returns new identity provider.
func New() bootstrap.IdentityProvider {
	return &jwtIDP{}
}

func (idp *jwtIDP) ExtractKey(token string) (string, error) {
	jwtToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	fmt.Printf("\n\nVALUE:%v\n", jwtToken.Claims)
	if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
		if id, ok := claims["sub"]; ok {
			return id.(string), nil
		}
	}

	return "", errInvalidJWTClaims
}

func (idp *jwtIDP) Identify(string) (string, error) {
	return "", nil
}
