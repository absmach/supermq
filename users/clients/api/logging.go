package api

import (
	"context"
	"fmt"
	"time"

	mflog "github.com/mainflux/mainflux/logger"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/users/clients"
	"github.com/mainflux/mainflux/users/jwt"
)

var _ clients.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger mflog.Logger
	svc    clients.Service
}

func LoggingMiddleware(svc clients.Service, logger mflog.Logger) clients.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) RegisterClient(ctx context.Context, token string, client mfclients.Client) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method register_client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.RegisterClient(ctx, token, client)
}

func (lm *loggingMiddleware) IssueToken(ctx context.Context, identity, secret string) (t jwt.Token, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method issue_token of type %s for client %s took %s to complete", t.AccessType, identity, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.IssueToken(ctx, identity, secret)
}

func (lm *loggingMiddleware) RefreshToken(ctx context.Context, refreshToken string) (t jwt.Token, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method refresh_token of type %s for token %s took %s to complete", t.AccessType, refreshToken, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.RefreshToken(ctx, refreshToken)
}

func (lm *loggingMiddleware) ViewClient(ctx context.Context, token, id string) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ViewClient(ctx, token, id)
}

func (lm *loggingMiddleware) ViewProfile(ctx context.Context, token string) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_profile with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ViewProfile(ctx, token)
}

func (lm *loggingMiddleware) ListClients(ctx context.Context, token string, pm mfclients.Page) (cp mfclients.ClientsPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_clients %d clients using token %s took %s to complete", cp.Total, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListClients(ctx, token, pm)
}

func (lm *loggingMiddleware) UpdateClient(ctx context.Context, token string, client mfclients.Client) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_name_and_metadata for client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClient(ctx, token, client)
}

func (lm *loggingMiddleware) UpdateClientTags(ctx context.Context, token string, client mfclients.Client) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_tags for client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientTags(ctx, token, client)
}
func (lm *loggingMiddleware) UpdateClientIdentity(ctx context.Context, token, id, identity string) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_identity for client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientIdentity(ctx, token, id, identity)
}

func (lm *loggingMiddleware) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_secret for client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)
}

func (lm *loggingMiddleware) GenerateResetToken(ctx context.Context, email, host string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method generate_reset_token for email %s took %s to complete", email, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.GenerateResetToken(ctx, email, host)
}

func (lm *loggingMiddleware) ResetSecret(ctx context.Context, token, secret string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method reset_secret using token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ResetSecret(ctx, token, secret)
}

func (lm *loggingMiddleware) SendPasswordReset(ctx context.Context, host, email, token string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method send_password_reset using token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.SendPasswordReset(ctx, host, email, token)
}

func (lm *loggingMiddleware) UpdateClientOwner(ctx context.Context, token string, client mfclients.Client) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_client_owner for client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdateClientOwner(ctx, token, client)
}

func (lm *loggingMiddleware) EnableClient(ctx context.Context, token, id string) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method enable_client for client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.EnableClient(ctx, token, id)
}

func (lm *loggingMiddleware) DisableClient(ctx context.Context, token, id string) (c mfclients.Client, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method disable_client for client with id %s using token %s took %s to complete", c.ID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.DisableClient(ctx, token, id)
}

func (lm *loggingMiddleware) ListMembers(ctx context.Context, token, groupID string, cp mfclients.Page) (mp mfclients.MembersPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_members %d members for group with id %s and token %s took %s to complete", mp.Total, groupID, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListMembers(ctx, token, groupID, cp)
}

func (lm *loggingMiddleware) Identify(ctx context.Context, token string) (id string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method identify for token %s with id %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.Identify(ctx, token)
}
