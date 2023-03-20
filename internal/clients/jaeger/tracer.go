// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package jaeger

import (
	"errors"

	jaegerp "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var (
	errNoUrl     = errors.New("URL is empty")
	errNoSvcName = errors.New("Service Name is empty")
)

// NewProvider initializes Jaeger TraceProvider
func NewProvider(svcName, url string) (*tracesdk.TracerProvider, error) {
	if url == "" {
		return nil, errNoUrl
	}

	if svcName == "" {
		return nil, errNoSvcName
	}

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(exporter)),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(svcName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(jaegerp.Jaeger{})

	return tp, nil
}
