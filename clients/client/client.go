// Package client provides a clients service client intended for internal
// service communication.
package client

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mainflux/mainflux/clients"
	"github.com/sony/gobreaker"
)

const (
	timeout         = time.Second * 5
	maxFailedReqs   = 3
	maxFailureRatio = 0.6
)

var (
	// ErrServiceUnreachable indicates that the service instance is not available.
	ErrServiceUnreachable = errors.New("clients service unavailable")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = clients.ErrUnauthorizedAccess
)

// ClientsClient provides an access to the clients service authorization
// endpoints.
type ClientsClient struct {
	url string
	cb  *gobreaker.CircuitBreaker
}

// NewClient instantiates the clients service client given its base URL.
func NewClient(url string) ClientsClient {
	st := gobreaker.Settings{
		Name: "Clients",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			fr := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= maxFailedReqs && fr >= maxFailureRatio
		},
	}

	mc := ClientsClient{
		url: url,
		cb:  gobreaker.NewCircuitBreaker(st),
	}

	return mc
}

// VerifyToken tries to extract an identity from the provided token.
func (mc ClientsClient) VerifyToken(token string) (string, error) {
	url := fmt.Sprintf("%s/access-grant", mc.url)
	return mc.makeRequest(url, token)
}

// CanAccess checks whether or not the client having a provided token has
// access to the specified channel.
func (mc ClientsClient) CanAccess(channel, token string) (string, error) {
	url := fmt.Sprintf("%s/channels/%s/access-grant", mc.url, channel)
	return mc.makeRequest(url, token)
}

func (mc ClientsClient) makeRequest(url, token string) (string, error) {
	response, err := mc.cb.Execute(func() (interface{}, error) {
		hc := &http.Client{
			Timeout: timeout,
		}

		mgReq, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", ErrServiceUnreachable
		}

		mgReq.Header.Set("Authorization", token)

		res, err := hc.Do(mgReq)
		if err != nil {
			return "", ErrServiceUnreachable
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return ErrUnauthorizedAccess, nil
		}

		return res.Header.Get("X-client-id"), nil
	})

	if err != nil {
		return "", err
	}

	id, ok := response.(string)
	if !ok {
		return "", clients.ErrUnauthorizedAccess
	}

	return id, nil
}
