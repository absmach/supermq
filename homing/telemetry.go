package homing

import "context"

type Telemetry struct {
	ID        string   `json:"-"`
	Services  []string `json:"-"`
	Longitude float64  `json:"-"`
	Latitutde float64  `json:"-"`
	IpAddress string   `json:"ip_address"`
	Version   string   `json:"mainflux_version"`
}

type PageMetadata struct {
	Total  uint64
	Offset uint64
	Limit  uint64
}

type TelemetryPage struct {
	PageMetadata
	Telemetry []Telemetry
}

// TelemetryRepository specifies an account persistence API.
type TelemetryRepo interface {
	// Save persists the telemetry event. A non-nil error is returned to indicate
	// operation failure.
	Save(ctx context.Context, t Telemetry) (string, error)

	// Update updates Telemetry event
	UpdateTelemetry(ctx context.Context, u Telemetry) error

	// RetrieveByIP retrieves telemetry by its unique identifier (i.e. ip address).
	RetrieveByIP(ctx context.Context, email string) (Telemetry, error)

	// RetrieveByID retrieves telemetry by its unique identifier ID.
	RetrieveByID(ctx context.Context, id string) (Telemetry, error)

	// RetrieveAll retrieves all telemetry for given array of telemetry IDs.
	RetrieveAll(ctx context.Context, IDs []string, pm PageMetadata) (Telemetry, error)
}
