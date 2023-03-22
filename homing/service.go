package homing

import "context"

type Service interface {
	Save(ctx context.Context, t Telemetry, serviceName string) error
	GetAll(ctx context.Context, token string, pm PageMetadata) (TelemetryPage, error)
}

var _ Service = (*telemetryService)(nil)

type telemetryService struct{}

// GetAll implements Service
func (ts *telemetryService) GetAll(ctx context.Context, token string, pm PageMetadata) (TelemetryPage, error) {
	panic("unimplemented")
}

// Save implements Service
func (ts *telemetryService) Save(ctx context.Context, t Telemetry, serviceName string) error {
	panic("unimplemented")
}
