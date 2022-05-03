package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	jose "github.com/go-jose/go-jose/v3"
	josejwt "github.com/go-jose/go-jose/v3/jwt"
	"github.com/golang-jwt/jwt/v4"
)

const (
	token    = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDg2MTcxMTEsImlhdCI6MTY0ODU4MTExMSwiaXNzIjoibWFpbmZsdXguYXV0aCIsInN1YiI6ImFkbWluQGV4YW1wbGUuY29tIiwiaXNzdWVyX2lkIjoiN2QwOTI0YmUtZTc2ZS00MWRiLWI4MGYtMmUyMzk1ZWFkN2I3IiwidHlwZSI6MH0.xr74BiNY0kBNENbA0ctH9AXbZgLa_MkvNrzp_6VpSU4`
	rawToken = `{
	"sub": "1234567890",
	"name": "John Doe",
	"iat": 1516239022
  }
`
)

var keyType = uint32(0)

type claims struct {
	jwt.StandardClaims
	IssuerID string  `json:"issuer_id,omitempty"`
	Type     *uint32 `json:"type,omitempty"`
}

func main() {
	b, err := ioutil.ReadFile("rsa.jwk")
	if err != nil {
		panic(err)
	}
	var jwk jose.JSONWebKey
	err = json.Unmarshal(b, &jwk)
	if err != nil {
		panic("error unmarshall")
	}

	fmt.Printf("jwk:\n%v\n", jwk.Use)
	signerKey := jose.SigningKey{
		Key: &jose.JSONWebKey{
			Key:   jwk.Key,
			KeyID: jwk.KeyID,
		},
		Algorithm: jose.SignatureAlgorithm(jwk.Algorithm),
	}

	sg, err := jose.NewSigner(signerKey, (&jose.SignerOptions{}).WithType("JWT").WithBase64(true))
	if err != nil {
		panic(err)
	}
	j, _ := sg.Sign([]byte(rawToken))

	fmt.Println(j.FullSerialize())

	claims := claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:   "issuerName",
			Subject:  "key.Subject",
			IssuedAt: time.Now().Unix(),
		},
		IssuerID: "key.IssuerID",
		Type:     &keyType,
	}

	signer, err := jose.NewSigner(signerKey, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		panic(err)
	}
	token, err := josejwt.Signed(signer).Claims(claims).CompactSerialize()
	fmt.Printf("token: %v \n", token)

	j

}
