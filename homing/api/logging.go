package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/homing"
	"github.com/mainflux/mainflux/logger"
)

var _ homing.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	hommingLogger logger.Logger
	svc           homing.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc homing.Service, logger logger.Logger) homing.Service {
	return &loggingMiddleware{logger, svc}
}

// GetAll implements homing.Service
func (lm *loggingMiddleware) GetAll(ctx context.Context, token string, pm homing.PageMetadata) (telemetryPage homing.TelemetryPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method get all telemetry took %s to complete", time.Since(begin))
		if err != nil {
			lm.hommingLogger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.hommingLogger.Info(fmt.Sprintf("%s without errors.", message))

	}(time.Now())

	return lm.svc.GetAll(ctx, token, pm)
}

// Save implements homing.Service
func (lm *loggingMiddleware) Save(ctx context.Context, t homing.Telemetry, serviceName string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method save telemetry event took %s to complete", time.Since(begin))
		if err != nil {
			lm.hommingLogger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.hommingLogger.Info(fmt.Sprintf("%s without errors.", message))

	}(time.Now())

	return lm.svc.Save(ctx, t, serviceName)
}
