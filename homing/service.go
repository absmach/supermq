package homing

import "context"

type Service interface {
	Save(ctx context.Context, t Telemetry, serviceName string) error
	GetAll(ctx context.Context, token string, pm PageMetadata) (TelemetryPage, error)
}
