//Package pwdrecovery provides password recovery token generation with TTL
//Token is sent by email to user as part of recovery URL
//Token contains 32 bytes where first 4 bytes are exparation time and 28 bytes random string
// exparation-email signed by secret signature
package pwdrecovery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Recovery token bytes
var (
	emailLength  = 254
	ttlLength    = 4
	secretLength = 20
	// Max token length is 4 bytes + max email length
	tokenLength = ttlLength + emailLength
	hashCost    = 10
	// Recovery token TTL in minutes, reperesents token time to live
	tokenDuration = 1
	// Random string for secret key, required for signing
	secret = "fcERNb7KpM3WyAmguJMZ"
	domain = "https://localhost"
	// Errors
	errMalformedToken  = errors.New("Malformed token")
	errExpiredToken    = errors.New("Token expired")
	errWrongSignature  = errors.New("Wrong token signature")
	errTokenGeneration = errors.New("Token generation faild")
)

// Generate NewToken generate new random token with defined TTL
func Generate(email string) (string, error) {

	b := make([]byte, ttlLength+len(email))
	expires := time.Now().Add(time.Minute * time.Duration(tokenDuration))
	// Put TTL and email time.
	binary.BigEndian.PutUint32(b, uint32(expires.Unix()))
	copy(b[ttlLength:], []byte(email))
	// hash the email part
	s, err := getSignature([]byte(b[ttlLength:]), []byte(secret))
	copy(b[ttlLength:], []byte(s))
	if err != nil {
		return "", errTokenGeneration
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Verify verifies token validity
func Verify(token string, hashed string) error {
	blen := base64.URLEncoding.DecodedLen(len(token))
	// Check max token length
	if blen > tokenLength {
		return errMalformedToken
	}
	b, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return errMalformedToken
	}
	// Compare token with stored hashed version
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(token)); err != nil {
		return errWrongSignature
	}
	// Verify exparation time
	ttl := time.Unix(int64(binary.BigEndian.Uint32(b[:ttlLength])), 0)
	if ttl.Before(time.Now()) {
		return errExpiredToken
	}
	return nil
}

// Hash hash's token value in order to save crypted token in DB
func Hash(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), hashCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Send sends password recovery link to user
func Send(email, token string) error {
	url := buildURL(token)
	fmt.Println("This is a recovery URL which I need to send", url, "On email", email)
	return nil
}

// Builds recovery URL
func buildURL(t string) string {
	return fmt.Sprintf("%s/password/recovery/token=%s", domain, t)
}

// Hashing token with signature
func getSignature(data []byte, signature []byte) ([]byte, error) {
	keym := hmac.New(sha256.New, signature)
	keym.Write(data)
	m := hmac.New(sha256.New, keym.Sum(nil))
	m.Write(data)
	return m.Sum(nil), nil
}
