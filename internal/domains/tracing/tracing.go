package tracing

import (
	"context"

	"github.com/absmach/magistrala/pkg/domains"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ domains.Service = (*tracingMiddleware)(nil)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    domains.Service
}

// New returns a new group service with tracing capabilities.
func New(svc domains.Service, tracer trace.Tracer) domains.Service {
	return &tracingMiddleware{tracer, svc}
}

func (tm *tracingMiddleware) CreateDomain(ctx context.Context, token string, d domains.Domain) (domains.Domain, error) {
	ctx, span := tm.tracer.Start(ctx, "create_domain", trace.WithAttributes(
		attribute.String("name", d.Name),
	))
	defer span.End()
	return tm.svc.CreateDomain(ctx, token, d)
}

func (tm *tracingMiddleware) RetrieveDomain(ctx context.Context, token, id string) (domains.Domain, error) {
	ctx, span := tm.tracer.Start(ctx, "view_domain", trace.WithAttributes(
		attribute.String("id", id),
	))
	defer span.End()
	return tm.svc.RetrieveDomain(ctx, token, id)
}

func (tm *tracingMiddleware) UpdateDomain(ctx context.Context, token, id string, d domains.DomainReq) (domains.Domain, error) {
	ctx, span := tm.tracer.Start(ctx, "update_domain", trace.WithAttributes(
		attribute.String("id", id),
	))
	defer span.End()
	return tm.svc.UpdateDomain(ctx, token, id, d)
}

func (tm *tracingMiddleware) ChangeDomainStatus(ctx context.Context, token, id string, d domains.DomainReq) (domains.Domain, error) {
	ctx, span := tm.tracer.Start(ctx, "change_domain_status", trace.WithAttributes(
		attribute.String("id", id),
	))
	defer span.End()
	return tm.svc.ChangeDomainStatus(ctx, token, id, d)
}

func (tm *tracingMiddleware) ListDomains(ctx context.Context, token string, p domains.Page) (domains.DomainsPage, error) {
	ctx, span := tm.tracer.Start(ctx, "list_domains")
	defer span.End()
	return tm.svc.ListDomains(ctx, token, p)
}
