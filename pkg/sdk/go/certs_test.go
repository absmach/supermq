// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/absmach/magistrala"
	authmocks "github.com/absmach/magistrala/auth/mocks"
	"github.com/absmach/magistrala/certs"
	"github.com/absmach/magistrala/certs/pki"
	httpapi "github.com/absmach/magistrala/certs/api"
	"github.com/absmach/magistrala/certs/mocks"
	"github.com/absmach/magistrala/internal/apiutil"
	mglog "github.com/absmach/magistrala/logger"
	"github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	sdk "github.com/absmach/magistrala/pkg/sdk/go"
	thmocks "github.com/absmach/magistrala/things/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const instanceID = "5de9b29a-feb9-11ed-be56-0242ac120002"

var thingID = "1"

var p = certs.Page{
	Total:  0,
	Offset: 0,
	Limit:  0,
	Certs:  []certs.Cert{},
}

func setupCerts() (*httptest.Server, *authmocks.AuthClient, *mocks.Agent, *mocks.Repository, *thmocks.Repository, error) {
	server, trepo, _, auth, _ := setupThings()
	config := sdk.Config{
		ThingsURL: server.URL,
	}

	mgsdk := sdk.NewSDK(config)

	repo := new(mocks.Repository)
	agent := new(mocks.Agent)

	svc := certs.New(auth, repo, mgsdk, agent)
	logger := mglog.NewMock()
	mux := httpapi.MakeHandler(svc, logger, instanceID)

	return httptest.NewServer(mux), auth, agent, repo, trepo, nil
}

func TestIssueCert(t *testing.T) {
	ts, auth, agent, repo, trepo, err := setupCerts()
	require.Nil(t, err, fmt.Sprintf("unexpected error during creating service: %s", err))
	defer ts.Close()

	sdkConf := sdk.Config{
		CertsURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc     string
		thingID  string
		duration string
		token    string
		err      errors.SDKError
	}{
		{
			desc:     "create new cert with thing id and duration",
			thingID:  thingID,
			duration: "10h",
			token:    validToken,
			err:      nil,
		},
		{
			desc:     "create new cert with empty thing id and duration",
			thingID:  "",
			duration: "10h",
			token:    validToken,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "create new cert with invalid thing id and duration",
			thingID:  "ah",
			duration: "10h",
			token:    validToken,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, certs.ErrFailedCertCreation), http.StatusInternalServerError),
		},
		{
			desc:     "create new cert with thing id and empty duration",
			thingID:  thingID,
			duration: "",
			token:    exampleUser1,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingCertData), http.StatusBadRequest),
		},
		{
			desc:     "create new cert with thing id and malformed duration",
			thingID:  thingID,
			duration: "10g",
			token:    exampleUser1,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidCertData), http.StatusBadRequest),
		},
		{
			desc:     "create new cert with empty token",
			thingID:  thingID,
			duration: "10h",
			token:    "",
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrBearerToken), http.StatusUnauthorized),
		},
		{
			desc:     "create new cert with invalid token",
			thingID:  thingID,
			duration: "10h",
			token:    authmocks.InvalidValue,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, svcerr.ErrAuthentication), http.StatusUnauthorized),
		},
		{
			desc:     "create new empty cert",
			thingID:  "",
			duration: "",
			token:    validToken,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
	}

	for _, tc := range cases {
		repoCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
		repoCall2 := trepo.On("RetrieveByID", mock.Anything, mock.Anything).Return(clients.Client{ID: tc.thingID}, tc.err)
		repoCall3 := agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, tc.err)
		repoCall4 := repo.On("Save", mock.Anything, mock.Anything).Return("", tc.err)

		cert, err := mgsdk.IssueCert(tc.thingID, tc.duration, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		if err == nil {
			assert.NotEmpty(t, cert, fmt.Sprintf("%s: got empty cert", tc.desc))
		}
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
		repoCall3.Unset()
		repoCall4.Unset()
	}
}

func TestViewCert(t *testing.T) {
	ts, auth, agent, repo, trepo, err := setupCerts()
	require.Nil(t, err, fmt.Sprintf("unexpected error during creating service: %s", err))
	defer ts.Close()

	c := certs.Cert{
		OwnerID:        "",
		ThingID:        thingID,
		ClientCert:     "",
		IssuingCA:      "",
		CAChain:        []string{},
		ClientKey:      "",
		PrivateKeyType: "",
		Serial:         "",
		Expire:         time.Time{},
	}

	sdkConf := sdk.Config{
		CertsURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	repoCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
	repoCall1 := auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	repoCall2 := trepo.On("RetrieveByID", mock.Anything, mock.Anything).Return(clients.Client{ID: thingID}, nil)
	repoCall3 := agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, nil)
	repoCall4 := repo.On("Save", mock.Anything, mock.Anything).Return("", nil)

	cert, err := mgsdk.IssueCert(thingID, "10h", token)
	require.Nil(t, err, fmt.Sprintf("unexpected error during creating cert: %s", err))
	fmt.Println("This is", cert.CertSerial)
	repoCall.Unset()
	repoCall1.Unset()
	repoCall2.Unset()
	repoCall3.Unset()
	repoCall4.Unset()

	cases := []struct {
		desc     string
		certID   string
		token    string
		err      errors.SDKError
		response sdk.Subscription
	}{
		{
			desc:     "get existing cert",
			certID:   validID,
			token:    token,
			err:      nil,
			response: sub1,
		},
		{
			desc:     "get non-existent cert",
			certID:   "43",
			token:    token,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, svcerr.ErrNotFound), http.StatusInternalServerError),
			response: sdk.Subscription{},
		},
		{
			desc:     "get cert with invalid token",
			certID:   validID,
			token:    "",
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrBearerToken), http.StatusUnauthorized),
			response: sdk.Subscription{},
		},
	}

	for _, tc := range cases {
		repoCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := repo.On("RetrieveBySerial", mock.Anything, mock.Anything, mock.Anything).Return(c, tc.err)
		repoCall2 := agent.On("Read", mock.Anything).Return(pki.Cert{}, tc.err)

		cert, err := mgsdk.ViewCert(tc.certID, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		if err == nil {
			assert.NotEmpty(t, cert, fmt.Sprintf("%s: got empty cert", tc.desc))
		}
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
	}
}

func TestViewCertByThing(t *testing.T) {
	ts, auth, agent, repo, trepo, err := setupCerts()
	require.Nil(t, err, fmt.Sprintf("unexpected error during creating service: %s", err))
	defer ts.Close()

	sdkConf := sdk.Config{
		CertsURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	repoCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
	repoCall1 := auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	repoCall2 := trepo.On("RetrieveByID", mock.Anything, mock.Anything).Return(clients.Client{ID: thingID}, nil)
	repoCall3 := agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, err)
	repoCall4 := repo.On("Save", mock.Anything, mock.Anything).Return("", err)

	_, err = mgsdk.IssueCert(thingID, "10h", token)
	require.Nil(t, err, fmt.Sprintf("unexpected error during creating cert: %s", err))
	repoCall.Unset()
	repoCall1.Unset()
	repoCall2.Unset()
	repoCall3.Unset()
	repoCall4.Unset()

	cases := []struct {
		desc     string
		thingID  string
		token    string
		err      errors.SDKError
		response sdk.Subscription
	}{
		{
			desc:     "get existing cert",
			thingID:  thingID,
			token:    token,
			err:      nil,
			response: sub1,
		},
		{
			desc:     "get non-existent cert",
			thingID:  "43",
			token:    token,
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, repoerr.ErrNotFound), http.StatusInternalServerError),
			response: sdk.Subscription{},
		},
		{
			desc:     "get cert with invalid token",
			thingID:  thingID,
			token:    "",
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrBearerToken), http.StatusUnauthorized),
			response: sdk.Subscription{},
		},
	}
	for _, tc := range cases {
		repoCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := repo.On("RetrieveByThing", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(p, tc.err)
		fmt.Println("This is the wrapped error1", tc.err)

		cert, err := mgsdk.ViewCertByThing(tc.thingID, tc.token)
		fmt.Println("This is the wrapped error2", err)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		if err == nil {
			assert.NotEmpty(t, cert, fmt.Sprintf("%s: got empty cert", tc.desc))
		}
		repoCall.Unset()
		repoCall1.Unset()
	}
}

// func TestRevokeCert(t *testing.T) {
// 	ts, auth, agent, repo, trepo, err := setupCerts()
// 	require.Nil(t, err, fmt.Sprintf("unexpected error during creating service: %s", err))
// 	defer ts.Close()

// 	sdkConf := sdk.Config{
// 		CertsURL:        ts.URL,
// 		MsgContentType:  contentType,
// 		TLSVerification: false,
// 	}

// 	mgsdk := sdk.NewSDK(sdkConf)

// 	repoCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
// 	repoCall1 := auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
// 	repoCall2 := trepo.On("RetrieveByID", mock.Anything, mock.Anything).Return(clients.Client{ID: thingID}, nil)
// 	repoCall3 := agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, err)
// 	repoCall4 := repo.On("Save", mock.Anything, mock.Anything).Return("", err)

// 	_, err = mgsdk.IssueCert(thingID, "10h", validToken)
// 	require.Nil(t, err, fmt.Sprintf("unexpected error during creating cert: %s", err))
// 	repoCall.Unset()
// 	repoCall1.Unset()
// 	repoCall2.Unset()
// 	repoCall3.Unset()
// 	repoCall4.Unset()

// 	cases := []struct {
// 		desc    string
// 		thingID string
// 		token   string
// 		err     errors.SDKError
// 	}{
// 		{
// 			desc:    "revoke cert with invalid token",
// 			thingID: thingID,
// 			token:   authmocks.InvalidValue,
// 			err:     errors.NewSDKErrorWithStatus(errors.Wrap(svcerr.ErrAuthentication, svcerr.ErrAuthentication), http.StatusUnauthorized),
// 		},
// 		{
// 			desc:    "revoke non-existing cert",
// 			thingID: "2",
// 			token:   token,
// 			err:     errors.NewSDKErrorWithStatus(errors.Wrap(svcerr.ErrNotFound, certs.ErrFailedCertRevocation), http.StatusInternalServerError),
// 		},
// 		{
// 			desc:    "revoke cert with empty token",
// 			thingID: thingID,
// 			token:   "",
// 			err:     errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrBearerToken, apiutil.ErrValidation), http.StatusUnauthorized),
// 		},
// 		{
// 			desc:    "revoke existing cert",
// 			thingID: thingID,
// 			token:   token,
// 			err:     nil,
// 		},
// 		{
// 			desc:    "revoke deleted cert",
// 			thingID: thingID,
// 			token:   token,
// 			err:     errors.NewSDKErrorWithStatus(errors.Wrap(svcerr.ErrNotFound, certs.ErrFailedToRemoveCertFromDB), http.StatusInternalServerError),
// 		},
// 	}

// 	for _, tc := range cases {
// 		repoCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
// 		repoCall1 := auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
// 		repoCall2 := trepo.On("RetrieveByID", mock.Anything, mock.Anything).Return(clients.Client{ID: tc.thingID}, nil)
// 		repoCall3 := repo.On("RetrieveByThing", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(p, tc.err)

// 		response, err := mgsdk.RevokeCert(tc.thingID, tc.token)
// 		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
// 		if err == nil {
// 			assert.NotEmpty(t, response, fmt.Sprintf("%s: got empty revocation time", tc.desc))
// 		}
// 		repoCall.Unset()
// 		repoCall1.Unset()
// 		repoCall2.Unset()
// 		repoCall3.Unset()
// 	}
// }
