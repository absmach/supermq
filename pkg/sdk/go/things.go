package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	thingsEndpoint     = "things"
	connectEndpoint    = "connect"
	disconnectEndpoint = "disconnect"
	identifyEndpoint   = "identify"
)

// Thing represents mainflux thing.
type Thing struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Credentials Credentials            `json:"credentials"`
	Tags        []string               `json:"tags,omitempty"`
	Owner       string                 `json:"owner,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Status      string                 `json:"status"`
}

// CreateThing creates a new client returning its id.
func (sdk mfSDK) CreateThing(thing Thing, token string) (Thing, errors.SDKError) {
	data, err := json.Marshal(thing)
	if err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.thingsURL, thingsEndpoint)

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusCreated)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	thing = Thing{}
	if err := json.Unmarshal(body, &thing); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return thing, nil
}

func (sdk mfSDK) CreateThings(things []Thing, token string) ([]Thing, errors.SDKError) {
	data, err := json.Marshal(things)
	if err != nil {
		return []Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, "bulk")

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), data, http.StatusCreated)
	if sdkerr != nil {
		return []Thing{}, sdkerr
	}

	var ctr createThingsRes
	if err := json.Unmarshal(body, &ctr); err != nil {
		return []Thing{}, errors.NewSDKError(err)
	}

	return ctr.Things, nil
}

// Things returns page of clients.
func (sdk mfSDK) Things(pm PageMetadata, token string) (ThingsPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.thingsURL, thingsEndpoint, pm)
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return ThingsPage{}, sdkerr
	}

	var cp ThingsPage
	if err := json.Unmarshal(body, &cp); err != nil {
		return ThingsPage{}, errors.NewSDKError(err)
	}

	return cp, nil
}

// ThingsByChannel retrieves everything that is assigned to a group identified by groupID.
func (sdk mfSDK) ThingsByChannel(chanID string, pm PageMetadata, token string) (ThingsPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.thingsURL, fmt.Sprintf("channels/%s/%s", chanID, thingsEndpoint), pm)
	if err != nil {
		return ThingsPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return ThingsPage{}, sdkerr
	}

	var tp ThingsPage
	if err := json.Unmarshal(body, &tp); err != nil {
		return ThingsPage{}, errors.NewSDKError(err)
	}

	return tp, nil
}

// Thing returns client object by id.
func (sdk mfSDK) Thing(id, token string) (Thing, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, id)

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	var t Thing
	if err := json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}

// UpdateThing updates existing client.
func (sdk mfSDK) UpdateThing(t Thing, token string) (Thing, errors.SDKError) {
	data, err := json.Marshal(t)
	if err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.thingsURL, thingsEndpoint, t.ID)

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	t = Thing{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}

// UpdateThingTags updates the client's tags.
func (sdk mfSDK) UpdateThingTags(t Thing, token string) (Thing, errors.SDKError) {
	data, err := json.Marshal(t)
	if err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/tags", sdk.thingsURL, thingsEndpoint, t.ID)

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	t = Thing{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}

// UpdateThingIdentity updates the thing's identity
func (sdk mfSDK) UpdateThingIdentity(t Thing, token string) (Thing, errors.SDKError) {
	ucir := updateClientIdentityReq{token: token, id: t.ID, Identity: t.Credentials.Identity}

	data, err := json.Marshal(ucir)
	if err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/identity", sdk.thingsURL, thingsEndpoint, t.ID)

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	t = Thing{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}

// UpdateThingSecret updates the client's secret
func (sdk mfSDK) UpdateThingSecret(id, oldSecret, newSecret, token string) (Thing, errors.SDKError) {
	var ucsr = updateClientSecretReq{OldSecret: oldSecret, NewSecret: newSecret}

	data, err := json.Marshal(ucsr)
	if err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/secret", sdk.thingsURL, thingsEndpoint, id)

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	var t Thing
	if err = json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}

// UpdateThingOwner updates the client's owner.
func (sdk mfSDK) UpdateThingOwner(t Thing, token string) (Thing, errors.SDKError) {
	data, err := json.Marshal(t)
	if err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/owner", sdk.thingsURL, thingsEndpoint, t.ID)

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, string(CTJSON), data, http.StatusOK)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	t = Thing{}
	if err = json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}

// EnableThing changes client status to enabled.
func (sdk mfSDK) EnableThing(id, token string) (Thing, errors.SDKError) {
	return sdk.changeThingStatus(id, enableEndpoint, token)
}

// DisableThing changes client status to disabled - soft delete.
func (sdk mfSDK) DisableThing(id, token string) (Thing, errors.SDKError) {
	return sdk.changeThingStatus(id, disableEndpoint, token)
}

func (sdk mfSDK) changeThingStatus(id, status, token string) (Thing, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.thingsURL, thingsEndpoint, id, status)
	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, string(CTJSON), nil, http.StatusOK)
	if sdkerr != nil {
		return Thing{}, sdkerr
	}

	t := Thing{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Thing{}, errors.NewSDKError(err)
	}

	return t, nil
}
