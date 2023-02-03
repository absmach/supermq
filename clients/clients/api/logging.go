package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/clients/clients"
	"github.com/mainflux/mainflux/clients/jwt"
	log "github.com/mainflux/mainflux/logger"
)

var _ clients.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    clients.Service
}

func LoggingMiddleware(svc clients.Service, logger log.Logger) clients.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) RegisterClient(ctx context.Context, token string, client clients.Client) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method register_client of identity %s with token %s took %s to complete", c.Credentials.Identity, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.RegisterClient(ctx, token, client)
}

func (lm *loggingMiddleware) IssueToken(ctx context.Context, identity, secret string) (token jwt.Token, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method issue_token for client %s took %s to complete", identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.IssueToken(ctx, identity, secret)
}

func (lm *loggingMiddleware) RefreshToken(ctx context.Context, accessToken string) (token jwt.Token, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method refresh_token for token %s took %s to complete", accessToken, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.RefreshToken(ctx, accessToken)
}

func (lm *loggingMiddleware) ViewClient(ctx context.Context, token, id string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_client for client %s took %s to complete", c.Credentials.Identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ViewClient(ctx, token, id)
}

func (lm *loggingMiddleware) ListClients(ctx context.Context, token string, pm clients.Page) (cp clients.ClientsPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_clients for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListClients(ctx, token, pm)
}

func (lm *loggingMiddleware) UpdateClient(ctx context.Context, token string, client clients.Client) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_name_and_metadata for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClient(ctx, token, client)
}

func (lm *loggingMiddleware) UpdateClientTags(ctx context.Context, token string, client clients.Client) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_tags for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientTags(ctx, token, client)
}
func (lm *loggingMiddleware) UpdateClientIdentity(ctx context.Context, token, id, identity string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_identity for token %s and identity %s took %s to complete", token, identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientIdentity(ctx, token, id, identity)
}

func (lm *loggingMiddleware) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_secret for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)
}

func (lm *loggingMiddleware) UpdateClientOwner(ctx context.Context, token string, client clients.Client) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_owner for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientOwner(ctx, token, client)
}

func (lm *loggingMiddleware) EnableClient(ctx context.Context, token string, id string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method enable_client for client %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.EnableClient(ctx, token, id)
}

func (lm *loggingMiddleware) DisableClient(ctx context.Context, token string, id string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method disable_client for client %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.DisableClient(ctx, token, id)
}

func (lm *loggingMiddleware) ListMembers(ctx context.Context, token, groupID string, cp clients.Page) (mp clients.MembersPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_members for group %s and token %s took %s to complete", groupID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListMembers(ctx, token, groupID, cp)
}
