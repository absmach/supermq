package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/provision"
	"github.com/mainflux/mainflux/provision/api"
	provSDK "github.com/mainflux/mainflux/provision/sdk"
	mfSDK "github.com/mainflux/mainflux/sdk/go"
)

const (
	defLogLevel        = "debug"
	defConfigFile      = "config.toml"
	defTLS             = "false"
	defCACerts         = ""
	defServerCert      = ""
	defServerKey       = ""
	defThingsLocation  = "http://localhost"
	defUsersLocation   = "http://localhost"
	defMQTTURL         = "localhost:1883"
	defHTTPPort        = "8091"
	defMfUser          = "test@example.com"
	defMfPass          = "test"
	defMfApiKey        = ""
	defMfBSURL         = "http://localhost:8202/things/configs"
	defMfWhiteListURL  = "http://localhost:8202/things/state"
	defMfCertsURL      = "http://localhost/certs"
	defProvisionCerts  = "false"
	defProvisionBS     = "true"
	defBSAutoWhitelist = "true"
	defBSContent       = `{}`

	envConfigFile       = "MF_PROVISION_CONFIG_FILE"
	envTLS              = "MF_PROVISION_ENV_CLIENTS_TLS"
	envCACerts          = "MFPROVISION_CA_CERTS"
	envServerCert       = "MF_PROVISION_SERVER_CERT"
	envServerKey        = "MF_PROVISION_SERVER_KEY"
	envMQTTURL          = "MF_PROVISION_MQTT_URL"
	envUsersLocation    = "MF_PROVISION_USERS_LOCATION"
	envThingsLocation   = "MF_PROVISION_THINGS_LOCATION"
	envLogLevel         = "MF_PROVISION_LOG_LEVEL"
	envHTTPPort         = "MF_PROVISION_HTTP_PORT"
	envMfUser           = "MF_PROVISION_USER"
	envMfPass           = "MF_PROVISION_PASS"
	envMfApiKey         = "MF_PROVISION_API_KEY"
	envMfBSURL          = "MF_PROVISION_BS_SVC_URL"
	envMfBSWhiteListURL = "MF_PROVISION_BS_SVC_WHITELIST_URL"
	envMfCertsURL       = "MF_PROVISION_CERTS_SVC_URL"
	envProvisionCerts   = "MF_PROVISION_X509_PROVISIONING"
	envProvisionBS      = "MF_PROVISION_BS_CONFIG_PROVISIONING"
	envBSAutoWhiteList  = "MF_PROVISION_BS_AUTO_WHITELIST"
	envBSContent        = "MF_PROVISION_BS_CONTENT"
)

var (
	errMissingConfigFile        = errors.New("missing config file setting")
	errFailedToGetAutoWhiteList = errors.New("failed to get auto whitelist setting")
	errGettingCertificate       = errors.New("getting certificate file setting")
	errGettingTLSConf           = errors.New("error getting certificate setting")
	errSettingProvBS            = errors.New("error getting provision BS url setting")
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf(err.Error())
	}
	logger, err := logger.New(os.Stdout, cfg.Server.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	cfg, err = loadConfigFromFile(cfg, logger)
	if err != nil {
		logger.Warn("Continue with settings from env vars")
	}

	thingSDKCfg := mfSDK.Config{
		BaseURL:           mainflux.Env(envThingsLocation, defThingsLocation),
		HTTPAdapterPrefix: "http",
		MsgContentType:    "application/json",
		TLSVerification:   cfg.Server.TLS,
	}
	thingsSDK := mfSDK.NewSDK(thingSDKCfg)

	userSDKCfg := mfSDK.Config{
		BaseURL:           mainflux.Env(envUsersLocation, defUsersLocation),
		HTTPAdapterPrefix: "http",
		MsgContentType:    "application/json",
		TLSVerification:   cfg.Server.TLS,
	}
	userSDK := mfSDK.NewSDK(userSDKCfg)
	SDK := provSDK.New(cfg.Server.MfCertsURL, cfg.Server.MfBSURL, cfg.Server.MfWhiteListURL, thingsSDK, userSDK)

	svc := provision.New(cfg, SDK, logger)
	svc = api.NewLoggingMiddleware(svc, logger)

	errs := make(chan error, 2)

	go startHTTPServer(svc, cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Provision service terminated: %s", err))
}

func startHTTPServer(svc provision.Service, cfg provision.Config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.Server.HTTPPort)
	if cfg.Server.ServerCert != "" || cfg.Server.ServerKey != "" {
		logger.Info(fmt.Sprintf("Provision service started using https on port %s with cert %s key %s",
			cfg.Server.HTTPPort, cfg.Server.ServerCert, cfg.Server.ServerKey))
		errs <- http.ListenAndServeTLS(p, cfg.Server.ServerCert, cfg.Server.ServerKey, api.MakeHandler(svc))
		return
	}
	logger.Info(fmt.Sprintf("Provision service started using http on port %s", cfg.Server.HTTPPort))
	errs <- http.ListenAndServe(p, api.MakeHandler(svc))
}

func loadConfigFromFile(cfg provision.Config, logger logger.Logger) (provision.Config, error) {
	_, err := os.Stat(cfg.File)
	if os.IsNotExist(err) {
		return cfg, errMissingConfigFile
	}
	c, err := provision.Read(cfg.File)
	if err != nil {
		return cfg, err
	}
	logger.Info("Continue with settings from file:" + cfg.File)
	return c, nil
}

func loadConfig() (provision.Config, error) {
	tls, err := strconv.ParseBool(mainflux.Env(envTLS, defTLS))
	if err != nil {
		return provision.Config{}, errors.Wrap(errGettingTLSConf, err)
	}
	provisionX509, err := strconv.ParseBool(mainflux.Env(envProvisionCerts, defProvisionCerts))
	if err != nil {
		return provision.Config{}, errors.Wrap(errGettingCertificate, err)
	}
	provisionBS, err := strconv.ParseBool(mainflux.Env(envProvisionBS, defProvisionBS))
	if err != nil {
		return provision.Config{}, errors.Wrap(errSettingProvBS, fmt.Errorf(" for %s", envProvisionBS))
	}

	autoWhiteList, err := strconv.ParseBool(mainflux.Env(envBSAutoWhiteList, defBSAutoWhitelist))
	if err != nil {
		return provision.Config{}, errors.Wrap(errFailedToGetAutoWhiteList, fmt.Errorf(" for %s", envBSAutoWhiteList))
	}
	if autoWhiteList && !provisionBS {
		return provision.Config{}, errors.New("Can't auto whitelist if auto config save is off")
	}

	cfg := provision.Config{
		Server: provision.ServiceConf{
			LogLevel:       mainflux.Env(envLogLevel, defLogLevel),
			CACerts:        mainflux.Env(envCACerts, defCACerts),
			ServerCert:     mainflux.Env(envServerCert, defServerCert),
			ServerKey:      mainflux.Env(envServerKey, defServerKey),
			HTTPPort:       mainflux.Env(envHTTPPort, defHTTPPort),
			MfBSURL:        mainflux.Env(envMfBSURL, defMfBSURL),
			MfWhiteListURL: mainflux.Env(envMfBSWhiteListURL, defMfWhiteListURL),
			MfCertsURL:     mainflux.Env(envMfCertsURL, defMfCertsURL),
			MfUser:         mainflux.Env(envMfUser, defMfUser),
			MfPass:         mainflux.Env(envMfPass, defMfPass),
			MfApiKey:       mainflux.Env(envMfApiKey, defMfApiKey),
			TLS:            tls,
		},
		Bootstrap: provision.Bootstrap{
			X509Provision: provisionX509,
			BSProvision:   provisionBS,
			AutoWhiteList: autoWhiteList,
			BSContent:     fmt.Sprintf(mainflux.Env(envBSContent, defBSContent), mainflux.Env(envMQTTURL, defMQTTURL)),
		},

		// This is default conf for provision if there is no
		// config file
		Channels: []provision.Channel{
			provision.Channel{
				Name:     "control-channel",
				Metadata: map[string]interface{}{"type": "control"},
			},
			provision.Channel{
				Name:     "data-channel",
				Metadata: map[string]interface{}{"type": "data"},
			},
		},
		Things: []provision.Thing{
			provision.Thing{
				Name:     "thing",
				Metadata: map[string]interface{}{"externalID": "xxxxxx"},
			},
		},
	}

	cfg.File = mainflux.Env(envConfigFile, "")
	if cfg.File == "" {
		cfg.File = defConfigFile
		if err != provision.Save(cfg, defConfigFile) {
			return cfg, err
		}
	}
	return cfg, nil
}
