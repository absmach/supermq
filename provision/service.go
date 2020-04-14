package provision

import (
	"fmt"

	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/logger"
	provSDK "github.com/mainflux/mainflux/provision/sdk"
	mfSDK "github.com/mainflux/mainflux/sdk/go"
)

const (
	ExternalID = "externalID"
	Active     = 1
)

var (
	errFailedToCreateToken = errors.New("failed to create access token")
	errEmptyThingsList     = errors.New("things list in configuration empty")
	errEmtpyChannelsList   = errors.New("channels list in configuration is empty")
	errMissingCredentials  = errors.New("missing credentials")
)

var _ Service = (*provisionService)(nil)

// Service specifies Provision service API.
type Service interface {
	// Provision is the only method this API specifies. Depending on the configuration,
	// the following actions will can be executed:
	// - create a Thing based od mac address
	// - create multiple Channels
	// - create Bootstrap configuration
	// - whitelist Thing in Bootstrap configuration == connect Thing to Channels
	Provision(token, externalID, externalKey string) (Result, error)
}

type provisionService struct {
	logger logger.Logger
	sdk    provSDK.SDK
	conf   Config
}

// Result represent what is created with additional info.
type Result struct {
	Things      []provSDK.Thing   `json:"things,omitempty"`
	ThingsID    []string          `json:"thing_ids,omitempty"`
	Channels    []provSDK.Channel `json:"channels,omitempty"`
	ClientCert  map[string]string `json:"client_cert,omitempty"`
	ClientKey   map[string]string `json:"client_key,omitempty"`
	CACert      string            `json:"ca_cert,omitempty"`
	Whitelisted map[string]bool   `json:"whitelisted,omitempty"`
}

// New returns new provision service.
func New(cfg Config, sdk provSDK.SDK, logger logger.Logger) Service {
	return &provisionService{
		logger: logger,
		conf:   cfg,
		sdk:    sdk,
	}
}

// Provision is provision method for adding devices to proxy.
func (ps *provisionService) Provision(token, externalID, externalKey string) (res Result, err error) {
	channels := make([]provSDK.Channel, 0)
	things := make([]provSDK.Thing, 0)
	defer ps.recover(&err, &things, &channels, &token)

	
	if token == "" {
		token = ps.conf.Server.MfApiKey
		if token == "" {
			if ps.conf.Server.MfUser == ""  || ps.conf.Server.MfPass == "" {
				return res, errMissingCredentials
			}
			token, err = ps.sdk.CreateToken(ps.conf.Server.MfUser, ps.conf.Server.MfPass)
			if err != nil {
				return res, errors.Wrap(errFailedToCreateToken, err)
			}
		}

	}
	if len(ps.conf.Things) == 0 {
		return res, errEmptyThingsList
	}
	if len(ps.conf.Channels) == 0 {
		return res, errEmtpyChannelsList
	}
	for _, thing := range ps.conf.Things {
		// If thing in configs contains metadata with externalid
		// set value for it from the provision request
		if _, ok := thing.Metadata[ExternalID]; ok {
			thing.Metadata[ExternalID] = externalID
		}
		th := mfSDK.Thing{
			Name:     thing.Name,
			Metadata: thing.Metadata,
		}
		thID, err := ps.sdk.CreateThing(th, token)
		if err != nil {
			return res, err
		}
		// Get newly created thing (in order to get the key).
		thing, err := ps.sdk.Thing(thID, token)
		if err != nil {
			return res, errors.Wrap(provSDK.ErrGetThing, fmt.Errorf("thing id:%s", thID))
		}
		things = append(things, thing)
	}

	for _, channel := range ps.conf.Channels {
		ch := mfSDK.Channel{
			Name:     channel.Name,
			Metadata: channel.Metadata,
		}
		chCreated, err := ps.sdk.CreateChannel(ch, token)
		if err != nil {
			return res, err
		}
		channels = append(channels, chCreated)
	}

	res = Result{
		Things:      things,
		Channels:    channels,
		Whitelisted: map[string]bool{},
		ClientCert:  map[string]string{},
		ClientKey:   map[string]string{},
	}

	var certs provSDK.Cert
	for _, thing := range things {
		if ps.conf.Bootstrap.X509Provision {
			certs, err = ps.sdk.Cert(thing.ID, thing.Key, token)
			if err != nil {
				return res, provSDK.ErrCerts
			}
			res.ClientCert[thing.ID] = certs.ClientCert
			res.ClientKey[thing.ID] = certs.ClientKey
			res.CACert = certs.CACert
		}
		chanIDs := []string{}
		for _, ch := range channels {
			chanIDs = append(chanIDs, ch.ID)
		}
		if ps.conf.Bootstrap.Provision {
			bsReq := provSDK.BSConfig{
				ThingID:     thing.ID,
				ExternalID:  externalID,
				ExternalKey: externalKey,
				Channels:    chanIDs,
				CACert:      res.CACert,
				ClientCert:  certs.ClientCert,
				ClientKey:   certs.ClientKey,
				Content:     ps.conf.Bootstrap.Content,
			}

			if err := ps.sdk.SaveConfig(bsReq, token); err != nil {
				return Result{}, errors.Wrap(provSDK.ErrConfig, err)
			}
		}

		if ps.conf.Bootstrap.AutoWhiteList {
			wlReq := map[string]int{
				"state": Active,
			}
			if err := ps.sdk.Whitelist(thing.ID, wlReq, token); err != nil {
				return res, provSDK.ErrWhitelist
			}
			res.Whitelisted[thing.ID] = true
		}
	}

	return res, nil
}

func (ps *provisionService) errLog(err error) {
	if err != nil {
		ps.logger.Error(fmt.Sprintf("Error recovering: %s", err))
	}
}

func clean(ps *provisionService, things []provSDK.Thing, channels []provSDK.Channel, token string) {
	for _, t := range things {
		ps.errLog(ps.sdk.DeleteThing(t.ID, token))
	}
	for _, c := range channels {
		ps.errLog(ps.sdk.DeleteThing(c.ID, token))
	}
}

func (ps *provisionService) recover(e *error, ths *[]provSDK.Thing, chs *[]provSDK.Channel, tkn *string) {
	things, channels, token, err := *ths, *chs, *tkn, *e
	switch err {
	case nil:
		return
	case provSDK.ErrGetThing, provSDK.ErrCreateCtrl:
		for _, th := range things {
			ps.errLog(ps.sdk.DeleteThing(th.ID, token))
		}
	case provSDK.ErrCreateData:
		for _, th := range things {
			ps.errLog(ps.sdk.DeleteThing(th.ID, token))
		}
		for _, ch := range channels {
			ps.errLog(ps.sdk.DeleteChannel(ch.ID, token))
		}
	case provSDK.ErrConn, provSDK.ErrCerts:
		clean(ps, things, channels, token)
	case provSDK.ErrConfig:
		clean(ps, things, channels, token)
		if ps.conf.Bootstrap.X509Provision {
			for _, th := range things {
				ps.errLog(ps.sdk.RemoveCert(th.ID, token))
			}
		}
	case provSDK.ErrWhitelist:
		clean(ps, things, channels, token)
		for _, th := range things {
			if ps.conf.Bootstrap.X509Provision {
				ps.errLog(ps.sdk.RemoveCert(th.ID, token))
			}
			ps.errLog(ps.sdk.RemoveConfig(th.ID, token))
		}
	}
}
