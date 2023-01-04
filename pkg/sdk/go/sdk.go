// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	// CTJSON represents JSON content type.
	CTJSON ContentType = "application/json"

	// CTJSONSenML represents JSON SenML content type.
	CTJSONSenML ContentType = "application/senml+json"

	// CTBinary represents binary content type.
	CTBinary ContentType = "application/octet-stream"
)

// ContentType represents all possible content types.
type ContentType string

var _ SDK = (*mfSDK)(nil)

// User represents mainflux user its credentials.
type User struct {
	ID       string                 `json:"id,omitempty"`
	Email    string                 `json:"email,omitempty"`
	Groups   []string               `json:"groups,omitempty"`
	Password string                 `json:"password,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
type PageMetadata struct {
	Total        uint64                 `json:"total"`
	Offset       uint64                 `json:"offset"`
	Limit        uint64                 `json:"limit"`
	Level        uint64                 `json:"level,omitempty"`
	Email        string                 `json:"email,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Disconnected bool                   `json:"disconnected,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Status       string                 `json:"status,omitempty"`
}

// Group represents mainflux users group.
type Group struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Thing represents mainflux thing.
type Thing struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Channel represents mainflux channel.
type Channel struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type Key struct {
	ID        string
	Type      uint32
	IssuerID  string
	Subject   string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// SDK contains Mainflux API.
type SDK interface {
	// CreateUser registers mainflux user.
	CreateUser(token string, user User) (string, errors.SDKError)

	// User returns user object by id.
	User(token, id string) (User, errors.SDKError)

	// Users returns list of users.
	Users(token string, pm PageMetadata) (UsersPage, errors.SDKError)

	// CreateToken receives credentials and returns user token.
	CreateToken(user User) (string, errors.SDKError)

	// UpdateUser updates existing user.
	UpdateUser(token string, user User) errors.SDKError

	// UpdatePassword updates user password.
	UpdatePassword(token, oldPass, newPass string) errors.SDKError

	// EnableUser changes the status of the user to enabled.
	EnableUser(id, token string) errors.SDKError

	// DisableUser changes the status of the user to disabled.
	DisableUser(id, token string) errors.SDKError

	// CreateThing registers new thing and returns its id.
	CreateThing(token string, thing Thing) (string, errors.SDKError)

	// CreateThings registers new things and returns their ids.
	CreateThings(token string, things []Thing) ([]Thing, errors.SDKError)

	// Things returns page of things.
	Things(token string, pm PageMetadata) (ThingsPage, errors.SDKError)

	// ThingsByChannel returns page of things that are connected or not connected
	// to specified channel.
	ThingsByChannel(token, chanID string, pm PageMetadata) (ThingsPage, errors.SDKError)

	// Thing returns thing object by id.
	Thing(token, id string) (Thing, errors.SDKError)

	// UpdateThing updates existing thing.
	UpdateThing(token string, thing Thing) errors.SDKError

	// DeleteThing removes existing thing.
	DeleteThing(token, id string) errors.SDKError

	// IdentifyThing validates thing's key and returns its ID
	IdentifyThing(key string) (string, errors.SDKError)

	// CreateGroup creates new group and returns its id.
	CreateGroup(token string, group Group) (string, errors.SDKError)

	// DeleteGroup deletes users group.
	DeleteGroup(token, id string) errors.SDKError

	// Groups returns page of groups.
	Groups(token string, pm PageMetadata) (GroupsPage, errors.SDKError)

	// Parents returns page of users groups.
	Parents(token, id string, pm PageMetadata) (GroupsPage, errors.SDKError)

	// Children returns page of users groups.
	Children(token, id string, pm PageMetadata) (GroupsPage, errors.SDKError)

	// Group returns users group object by id.
	Group(token, id string) (Group, errors.SDKError)

	// Assign assigns member of member type (thing or user) to a group.
	Assign(token string, memberIDs []string, memberType, groupID string) errors.SDKError

	// Unassign removes member from a group.
	Unassign(token, groupID string, memberIDs ...string) errors.SDKError

	// Members lists members of a group.
	Members(token, groupID string, pm PageMetadata) (MembersPage, errors.SDKError)

	// Memberships lists groups for user.
	Memberships(token, userID string, pm PageMetadata) (GroupsPage, errors.SDKError)

	// UpdateGroup updates existing group.
	UpdateGroup(token string, group Group) errors.SDKError

	// Connect bulk connects things to channels specified by id.
	Connect(token string, conns ConnectionIDs) errors.SDKError

	// DisconnectThing disconnect thing from specified channel by id.
	DisconnectThing(token, thingID, chanID string) errors.SDKError

	// CreateChannel creates new channel and returns its id.
	CreateChannel(token string, channel Channel) (string, errors.SDKError)

	// CreateChannels registers new channels and returns their ids.
	CreateChannels(token string, channels []Channel) ([]Channel, errors.SDKError)

	// Channels returns page of channels.
	Channels(token string, pm PageMetadata) (ChannelsPage, errors.SDKError)

	// ChannelsByThing returns page of channels that are connected or not connected
	// to specified thing.
	ChannelsByThing(token, thingID string, pm PageMetadata) (ChannelsPage, errors.SDKError)

	// Channel returns channel data by id.
	Channel(token, id string) (Channel, errors.SDKError)

	// UpdateChannel updates existing channel.
	UpdateChannel(token string, channel Channel) errors.SDKError

	// DeleteChannel removes existing channel.
	DeleteChannel(token, id string) errors.SDKError

	// SendMessage send message to specified channel.
	SendMessage(token, chanID, msg string) errors.SDKError

	// ReadMessages read messages of specified channel.
	ReadMessages(token, chanID string) (MessagesPage, errors.SDKError)

	// SetContentType sets message content type.
	SetContentType(ct ContentType) errors.SDKError

	// Health returns things service health check.
	Health() (mainflux.HealthInfo, errors.SDKError)

	// AddBootstrap add bootstrap configuration
	AddBootstrap(token string, cfg BootstrapConfig) (string, errors.SDKError)

	// View returns Thing Config with given ID belonging to the user identified by the given token.
	ViewBootstrap(token, id string) (BootstrapConfig, errors.SDKError)

	// Update updates editable fields of the provided Config.
	UpdateBootstrap(token string, cfg BootstrapConfig) errors.SDKError

	// Update boostrap config certificates
	UpdateBootstrapCerts(token string, id string, clientCert, clientKey, ca string) errors.SDKError

	// Remove removes Config with specified token that belongs to the user identified by the given token.
	RemoveBootstrap(token, id string) errors.SDKError

	// Bootstrap returns Config to the Thing with provided external ID using external key.
	Bootstrap(externalKey, externalID string) (BootstrapConfig, errors.SDKError)

	// Whitelist updates Thing state Config with given ID belonging to the user identified by the given token.
	Whitelist(token string, cfg BootstrapConfig) errors.SDKError

	// IssueCert issues a certificate for a thing required for mtls.
	IssueCert(token, thingID string, keyBits int, keyType, valid string) (Cert, errors.SDKError)

	// RemoveCert removes a certificate
	RemoveCert(token, id string) errors.SDKError

	// RevokeCert revokes certificate with certID for thing with thingID
	RevokeCert(token, thingID, certID string) errors.SDKError

	// Issue issues a new key, returning its token value alongside.
	Issue(token string, duration time.Duration) (KeyRes, errors.SDKError)

	// Revoke removes the key with the provided ID that is issued by the user identified by the provided key.
	Revoke(token, id string) errors.SDKError

	// RetrieveKey retrieves data for the key identified by the provided ID, that is issued by the user identified by the provided key.
	RetrieveKey(token, id string) (retrieveKeyRes, errors.SDKError)
}

type mfSDK struct {
	authURL        string
	bootstrapURL   string
	certsURL       string
	httpAdapterURL string
	readerURL      string
	thingsURL      string
	usersURL       string

	msgContentType ContentType
	client         *http.Client
}

// Config contains sdk configuration parameters.
type Config struct {
	AuthURL        string
	BootstrapURL   string
	CertsURL       string
	HTTPAdapterURL string
	ReaderURL      string
	ThingsURL      string
	UsersURL       string

	MsgContentType  ContentType
	TLSVerification bool
}

// NewSDK returns new mainflux SDK instance.
func NewSDK(conf Config) SDK {
	return &mfSDK{
		authURL:        conf.AuthURL,
		bootstrapURL:   conf.BootstrapURL,
		certsURL:       conf.CertsURL,
		httpAdapterURL: conf.HTTPAdapterURL,
		readerURL:      conf.ReaderURL,
		thingsURL:      conf.ThingsURL,
		usersURL:       conf.UsersURL,

		msgContentType: conf.MsgContentType,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !conf.TLSVerification,
				},
			},
		},
	}
}

// processRequest creates and send a new HTTP request, and checks for errors in the HTTP response.
// It then returns the response headers, the response body, and the associated error(s) (if any).
func (sdk mfSDK) processRequest(method, url, token, contentType string, data []byte, expectedRespCodes ...int) (http.Header, []byte, errors.SDKError) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}

	if token != "" {
		if !strings.Contains(token, apiutil.ThingPrefix) {
			token = apiutil.BearerPrefix + token
		}
		req.Header.Set("Authorization", token)
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	resp, err := sdk.client.Do(req)
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}
	defer resp.Body.Close()

	sdkerr := errors.CheckError(resp, expectedRespCodes...)
	if sdkerr != nil {
		return make(http.Header), []byte{}, sdkerr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make(http.Header), []byte{}, errors.NewSDKError(err)
	}

	return resp.Header, body, nil
}

func (sdk mfSDK) withQueryParams(baseURL, endpoint string, pm PageMetadata) (string, error) {
	q, err := pm.query()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s?%s", baseURL, endpoint, q), nil
}

func (pm PageMetadata) query() (string, error) {
	q := url.Values{}
	q.Add("total", strconv.FormatUint(pm.Total, 10))
	q.Add("offset", strconv.FormatUint(pm.Offset, 10))
	q.Add("limit", strconv.FormatUint(pm.Limit, 10))
	q.Add("disconnected", strconv.FormatBool(pm.Disconnected))
	if pm.Level != 0 {
		q.Add("level", strconv.FormatUint(pm.Level, 10))
	}
	if pm.Email != "" {
		q.Add("email", pm.Email)
	}
	if pm.Name != "" {
		q.Add("name", pm.Name)
	}
	if pm.Type != "" {
		q.Add("type", pm.Type)
	}
	if pm.Status != "" {
		q.Add("status", pm.Status)
	}
	if pm.Metadata != nil {
		md, err := json.Marshal(pm.Metadata)
		if err != nil {
			return "", errors.NewSDKError(err)
		}
		q.Add("metadata", string(md))
	}
	return q.Encode(), nil
}
