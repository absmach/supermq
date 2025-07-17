// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package openbao wraps OpenBao client for PKI operations
package openbao

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/absmach/supermq/certs"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/openbao/openbao/api/v2"
)

const (
	issue  = "issue"
	cert   = "cert"
	revoke = "revoke"
)

var (
	errFailedCertDecoding = errors.New("failed to decode response from OpenBao service")
	errFailedToLogin      = errors.New("failed to login to OpenBao")
	errFailedAppRole      = errors.New("failed to create OpenBao new app role")
	errNoAuthInfo         = errors.New("no auth information from OpenBao")
	errNonRenewal         = errors.New("token is not configured to be renewable")
	errRenewWatcher       = errors.New("unable to initialize new lifetime watcher for renewing auth token")
	errFailedRenew        = errors.New("failed to renew token")
	errCouldNotRenew      = errors.New("token can no longer be renewed")
)

// Agent represents the OpenBao PKI interface
type Agent interface {
	Issue(entityId, ttl string, ipAddrs []string) (certs.Cert, error)
	View(serialNumber string) (certs.Cert, error)
	Revoke(serialNumber string) error
	ListCerts(pm certs.PageMetadata) (certs.CertPage, error)
}

type openbaoPKIAgent struct {
	appRole   string
	appSecret string
	namespace string
	path      string
	role      string
	host      string
	issueURL  string
	readURL   string
	revokeURL string
	client    *api.Client
	secret    *api.Secret
	logger    *slog.Logger
}

type certReq struct {
	CommonName string `json:"common_name"`
	TTL        string `json:"ttl"`
}

type certRevokeReq struct {
	SerialNumber string `json:"serial_number"`
}

// NewAgent instantiates an OpenBao PKI client
func NewAgent(appRole, appSecret, host, namespace, path, role string, logger *slog.Logger) (Agent, error) {
	conf := api.DefaultConfig()
	conf.Address = host

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	if namespace != "" {
		client.SetNamespace(namespace)
	}

	p := openbaoPKIAgent{
		appRole:   appRole,
		appSecret: appSecret,
		host:      host,
		namespace: namespace,
		role:      role,
		path:      path,
		client:    client,
		logger:    logger,
		issueURL:  "/" + path + "/" + issue + "/" + role,
		readURL:   "/" + path + "/" + cert + "/",
		revokeURL: "/" + path + "/" + revoke,
	}
	return &p, nil
}

func (va *openbaoPKIAgent) Issue(entityId, ttl string, ipAddrs []string) (certs.Cert, error) {
	err := va.LoginAndRenew()
	if err != nil {
		return certs.Cert{}, err
	}

	secretValues := map[string]interface{}{
		"common_name":          entityId,
		"ttl":                  ttl,
		"ip_sans":              ipAddrs,
		"exclude_cn_from_sans": true,
	}

	if len(ipAddrs) > 0 {
		secretValues["ip_sans"] = ipAddrs
	}

	secret, err := va.client.Logical().Write(va.path+"/issue/"+va.role, secretValues)
	if err != nil {
		return certs.Cert{}, err
	}

	if secret == nil || secret.Data == nil {
		return certs.Cert{}, fmt.Errorf("no certificate data returned from OpenBao")
	}

	cert := certs.Cert{
		ClientID: entityId,
	}

	if certData, ok := secret.Data["certificate"].(string); ok {
		cert.Certificate = certData
	}

	if keyData, ok := secret.Data["private_key"].(string); ok {
		cert.Key = keyData
	}

	if serialNumber, ok := secret.Data["serial_number"].(string); ok {
		cert.SerialNumber = serialNumber
	}
	if caChain, ok := secret.Data["ca_chain"].([]interface{}); ok {
		for _, ca := range caChain {
			if caStr, ok := ca.(string); ok {
				cert.CAChain = append(cert.CAChain, caStr)
			}
		}
	}
	if issuingCA, ok := secret.Data["issuing_ca"].(string); ok {
		cert.IssuingCA = issuingCA
	}

	if expirationInterface, ok := secret.Data["expiration"]; ok {
		switch exp := expirationInterface.(type) {
		case int64:
			cert.ExpiryTime = time.Unix(exp, 0)
		case float64:
			cert.ExpiryTime = time.Unix(int64(exp), 0)
		case json.Number:
			if expInt, err := exp.Int64(); err == nil {
				cert.ExpiryTime = time.Unix(expInt, 0)
			}
		}
	}

	return cert, nil
}

func (va *openbaoPKIAgent) View(serialNumber string) (certs.Cert, error) {
	err := va.LoginAndRenew()
	if err != nil {
		return certs.Cert{}, err
	}

	secret, err := va.client.Logical().Read(va.path + "/cert/" + serialNumber)
	if err != nil {
		return certs.Cert{}, err
	}

	if secret == nil || secret.Data == nil {
		return certs.Cert{}, fmt.Errorf("certificate not found")
	}

	cert := certs.Cert{
		SerialNumber: serialNumber,
	}

	if certData, ok := secret.Data["certificate"].(string); ok {
		cert.Certificate = certData
	}

	return cert, nil
}

func (va *openbaoPKIAgent) Revoke(serialNumber string) error {
	err := va.LoginAndRenew()
	if err != nil {
		return err
	}

	secretValues := map[string]interface{}{
		"serial_number": serialNumber,
	}

	_, err = va.client.Logical().Write(va.path+"/revoke", secretValues)
	if err != nil {
		return err
	}

	return nil
}

func (va *openbaoPKIAgent) ListCerts(pm certs.PageMetadata) (certs.CertPage, error) {
	err := va.LoginAndRenew()
	if err != nil {
		return certs.CertPage{}, err
	}

	secret, err := va.client.Logical().List(va.path + "/certs")
	if err != nil {
		return certs.CertPage{}, err
	}

	certPage := certs.CertPage{
		Certificates: []certs.Cert{},
		Limit:        pm.Limit,
		Offset:       pm.Offset,
	}

	if secret == nil || secret.Data == nil {
		return certPage, nil
	}

	keysInterface, ok := secret.Data["keys"]
	if !ok {
		return certPage, nil
	}

	var serialNumbers []string
	if err := mapstructure.Decode(keysInterface, &serialNumbers); err != nil {
		return certPage, fmt.Errorf("failed to decode certificate serial numbers: %w", err)
	}

	certPage.Total = uint64(len(serialNumbers))

	start := pm.Offset
	end := pm.Offset + pm.Limit
	if start >= uint64(len(serialNumbers)) {
		return certPage, nil
	}
	if end > uint64(len(serialNumbers)) {
		end = uint64(len(serialNumbers))
	}

	for i := start; i < end; i++ {
		serialNumber := serialNumbers[i]
		cert, err := va.View(serialNumber)
		if err != nil {
			va.logger.Warn("failed to retrieve certificate details", "serial", serialNumber, "error", err)
			continue
		}
		certPage.Certificates = append(certPage.Certificates, cert)
	}

	return certPage, nil
}

func (va *openbaoPKIAgent) LoginAndRenew() error {
	if va.secret != nil && va.secret.Auth != nil && va.secret.Auth.ClientToken != "" {
		_, err := va.client.Auth().Token().LookupSelf()
		if err == nil {
			return nil
		}
	}

	authData := map[string]interface{}{
		"role_id":   va.appRole,
		"secret_id": va.appSecret,
	}

	authResp, err := va.client.Logical().Write("auth/approle/login", authData)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToLogin, err)
	}

	if authResp == nil || authResp.Auth == nil {
		return errNoAuthInfo
	}

	va.secret = authResp
	va.client.SetToken(authResp.Auth.ClientToken)

	if authResp.Auth.Renewable {
		watcher, err := va.client.NewLifetimeWatcher(&api.LifetimeWatcherInput{
			Secret: authResp,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", errRenewWatcher, err)
		}

		go va.renewToken(watcher)
	}

	return nil
}

func (va *openbaoPKIAgent) renewToken(watcher *api.LifetimeWatcher) {
	defer watcher.Stop()

	watcher.Start()
	for {
		select {
		case err := <-watcher.DoneCh():
			if err != nil {
				va.logger.Error("token renewal failed", "error", err)
			}
			return
		case renewal := <-watcher.RenewCh():
			va.logger.Info("token renewed successfully", "lease_duration", renewal.Secret.LeaseDuration)
		}
	}
}
