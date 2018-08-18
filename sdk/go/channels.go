package sdk

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const channelsEndpoint = "channels"

// CreateChannel - creates new channel and generates UUID
func CreateChannel(data, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", serverAddr, channelsEndpoint)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// GetChannels - gets all channels
func GetChannels(token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s?offset=%s&limit=%s",
		serverAddr, channelsEndpoint, strconv.Itoa(offset), strconv.Itoa(limit))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// GetChannel - gets channel by ID
func GetChannel(id, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, channelsEndpoint, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// UpdateChannel - update a channel
func UpdateChannel(id, data, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, channelsEndpoint, id)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}

// DeleteChannel - removes channel
func DeleteChannel(id, token string) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/%s", serverAddr, channelsEndpoint, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, contentTypeJSON)
}
