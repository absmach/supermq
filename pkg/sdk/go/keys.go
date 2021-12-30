// package sdk

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"

// 	"github.com/mainflux/mainflux/pkg/errors"
// )

// const keysEndpoint = "keys"

// func (sdk mfSDK) Issue(token string, k Key) (keyResponse, error) {
// 	data, err := json.Marshal(k)
// 	if err != nil {
// 		return Key{}, err
// 	}

// 	url := fmt.Sprintf("%s/%s", sdk.authURL, keysEndpoint)

// 	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
// 	if err != nil {
// 		return Key{}, err
// 	}

// 	resp, err := sdk.sendRequest(req, token, string(CTJSON))
// 	if err != nil {
// 		return Key{}, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return Key{}, err
// 	}

// 	if resp.StatusCode != http.StatusCreated {
// 		return Key{}, errors.Wrap(ErrFailedCreation, errors.New(resp.Status))
// 	}

// 	var k Key
// 	if err := json.Unmarshal(body, &k); err != nil {
// 		return Key{}, err
// 	}

// 	// return k, k.value, nil
// }

// func (sdk mfSDK) Revoke(id, token string) error {
// 	url := fmt.Sprintf("%s/%s/%s", sdk.authURL, keysEndpoint, id)
// 	req, err := http.NewRequest(http.MethodDelete, url, nil)
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := sdk.sendRequest(req, token, string(CTJSON))
// 	if err != nil {
// 		return err
// 	}

// 	if resp.StatusCode != http.StatusNoContent {
// 		return errors.Wrap(ErrFailedRemoval, errors.New(resp.Status))
// 	}

// 	return nil
// }

// func (sdk mfSDK) RetrieveKey(id, token string) (Key, error) {
// 	url := fmt.Sprintf("%s/%s/%s", sdk.authURL, keysEndpoint, id)
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		return Key{}, err
// 	}

// 	resp, err := sdk.sendRequest(req, token, string(CTJSON))
// 	if err != nil {
// 		return Key{}, err
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return Key{}, err
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return Key{}, errors.Wrap(ErrFailedFetch, errors.New(resp.Status))
// 	}

// 	var k Key
// 	if err := json.Unmarshal(body, &k); err != nil {
// 		return Key{}, err
// 	}

// 	return k, nil
// }