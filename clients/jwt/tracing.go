package jwt

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ TokenRepository = (*tokenRepoMiddlware)(nil)

type tokenRepoMiddlware struct {
	repo   TokenRepository
	tracer trace.Tracer
}

// NewTokenRepoMiddleware instantiates an implementation of tracing Token repository.
func NewTokenRepoMiddleware(repo TokenRepository, tracer trace.Tracer) TokenRepository {
	return &tokenRepoMiddlware{
		repo:   repo,
		tracer: tracer,
	}
}

func (trm tokenRepoMiddlware) Issue(ctx context.Context, claim Claims) (Token, error) {
	ctx, span := trm.tracer.Start(ctx, "issue_token", trace.WithAttributes(attribute.String("clientid", claim.ClientID)))
	defer span.End()

	return trm.repo.Issue(ctx, claim)
}

func (trm tokenRepoMiddlware) Parse(ctx context.Context, accessToken string) (Claims, error) {
	ctx, span := trm.tracer.Start(ctx, "parse_token", trace.WithAttributes(attribute.String("accesstoken", accessToken)))
	defer span.End()

	return trm.repo.Parse(ctx, accessToken)
}
