package sdk

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	thingsEndpoint = "things"
)

// CreateThing - creates new thing and generates thing UUID
func CreateThing(data, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", serverAddr, thingsEndpoint)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// GetThings - gets all things
func GetThings(token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s?offset=%s&limit=%s",
		serverAddr, thingsEndpoint, strconv.Itoa(offset), strconv.Itoa(limit))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// GetThing - gets thing by ID
func GetThing(id, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEndpoint, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// UpdateThing - updates thing by ID
func UpdateThing(id, data, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEndpoint, id)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// DeleteThing - removes thing
func DeleteThing(id, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, thingsEndpoint, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// ConnectThing - connect thing to a channel
func ConnectThing(cliId, chanId, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, channelsEndpoint, chanId, thingsEndpoint, cliId)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// DisconnectThing - connect thing to a channel
func DisconnectThing(cliId, chanId, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", serverAddr, channelsEndpoint, chanId, thingsEndpoint, cliId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}
