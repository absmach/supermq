package homing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	HomeUrl = "localhost:9022"
)

func New(svc, version string, homingLogger logger.Logger) *homingService {
	return &homingService{
		serviceName:  svc,
		version:      version,
		homingLogger: homingLogger,
	}
}

type homingService struct {
	serviceName  string
	version      string
	homingLogger logger.Logger
}

func (hs *homingService) CallHome(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var data telemetryData
			data.Service = hs.serviceName
			data.Version = hs.version
			ip, err := getIP()
			if err != nil {
				hs.homingLogger.Warn(errors.Wrap(fmt.Errorf("error getting ip address"), err).Error())
				continue
			}
			data.IpAddress = ip
			if err = sendTelemetry(&data); err != nil {
				hs.homingLogger.Warn(errors.Wrap(fmt.Errorf("error sending telemtry data"), err).Error())
				continue
			}
		}
		time.Sleep(time.Hour * 2)
	}
}

type telemetryData struct {
	Service   string `json:"service"`
	IpAddress string `json:"ip_address"`
	Version   string `json:"mainflux_version"`
}

func getIP() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://checkip.amazonaws.com/", nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func sendTelemetry(telDat *telemetryData) error {
	b, err := json.Marshal(telDat)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, HomeUrl, bytes.NewReader(b))
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	return err
}
