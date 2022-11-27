package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type keyReq struct {
	Type     uint32        `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

const keysEndpoint = "keys"

const (
	// LoginKey is temporary User key received on successfull login.
	LoginKey uint32 = iota
	// RecoveryKey represents a key for resseting password.
	RecoveryKey
	// APIKey enables the one to act on behalf of the user.
	APIKey
)

func (sdk mfSDK) Issue(token string, d time.Duration) (KeyRes, error) {
	datareq := keyReq{Type: APIKey, Duration: d}
	data, err := json.Marshal(datareq)
	if err != nil {
		return KeyRes{}, err
	}

	url := fmt.Sprintf("%s/%s", sdk.authURL, keysEndpoint)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return KeyRes{}, err
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return KeyRes{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return KeyRes{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		return KeyRes{}, encodeError(body, resp.StatusCode)
	}

	var key KeyRes
	if err := json.Unmarshal(body, &key); err != nil {
		return KeyRes{}, err
	}

	return key, nil
}

func (sdk mfSDK) Revoke(id, token string) error {
	url := fmt.Sprintf("%s/%s/%s", sdk.authURL, keysEndpoint, id)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return encodeError(body, resp.StatusCode)
	}

	return nil
}

func (sdk mfSDK) RetrieveKey(id, token string) (retrieveKeyRes, error) {
	url := fmt.Sprintf("%s/%s/%s", sdk.authURL, keysEndpoint, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return retrieveKeyRes{}, err
	}

	resp, err := sdk.sendRequest(req, token, string(CTJSON))
	if err != nil {
		return retrieveKeyRes{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return retrieveKeyRes{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return retrieveKeyRes{}, encodeError(body, resp.StatusCode)
	}

	var key retrieveKeyRes
	if err := json.Unmarshal(body, &key); err != nil {
		return retrieveKeyRes{}, err
	}

	return key, nil
}
