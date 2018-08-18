package sdk

import (
	"errors"
	"net/http"
	"strings"
)

// Default msgContentType is SenML
var msgContentType = contentTypeSenMLJSON

// SendMessage - send message on Mainflux channel
func SendMessage(id, msg, token string) (*http.Response, error) {
	url := serverAddr + "/http/channels/" + id + "/messages"

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(msg))
	if err != nil {
		return nil, err
	}

	return sendRequest(req, token, msgContentType)
}

// SetContentType - set message content type.
// Available options are SenML JSON, custom JSON and custom binary (octet-stream).
func SetContentType(ct string) error {
	if ct != contentTypeJSON && ct != contentTypeSenMLJSON && ct != contentTypeBinary {
		return errors.New("Unknown Content Type")
	}

	msgContentType = ct

	return nil
}
