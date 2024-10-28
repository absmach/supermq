// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/absmach/magistrala/pat"
)

var _ pat.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger *slog.Logger
	svc    pat.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc pat.Service, logger *slog.Logger) pat.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) CreatePAT(ctx context.Context, token, name, description string, duration time.Duration, scope pat.Scope) (pa pat.PAT, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("name", name),
			slog.String("description", description),
			slog.String("pat_duration", duration.String()),
			slog.String("scope", scope.String()),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Create PAT failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Create PAT completed successfully", args...)
	}(time.Now())
	return lm.svc.CreatePAT(ctx, token, name, description, duration, scope)
}

func (lm *loggingMiddleware) UpdatePATName(ctx context.Context, token, patID, name string) (pa pat.PAT, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
			slog.String("name", name),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Update PAT name failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Update PAT name completed successfully", args...)
	}(time.Now())
	return lm.svc.UpdatePATName(ctx, token, patID, name)
}

func (lm *loggingMiddleware) UpdatePATDescription(ctx context.Context, token, patID, description string) (pa pat.PAT, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
			slog.String("description", description),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Update PAT description failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Update PAT description completed successfully", args...)
	}(time.Now())
	return lm.svc.UpdatePATDescription(ctx, token, patID, description)
}

func (lm *loggingMiddleware) RetrievePAT(ctx context.Context, token, patID string) (pa pat.PAT, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Retrieve PAT  failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Retrieve PAT completed successfully", args...)
	}(time.Now())
	return lm.svc.RetrievePAT(ctx, token, patID)
}

func (lm *loggingMiddleware) ListPATS(ctx context.Context, token string, pm pat.PATSPageMeta) (pp pat.PATSPage, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Uint64("limit", pm.Limit),
			slog.Uint64("offset", pm.Offset),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("List PATS  failed to complete successfully", args...)
			return
		}
		lm.logger.Info("List PATS completed successfully", args...)
	}(time.Now())
	return lm.svc.ListPATS(ctx, token, pm)
}

func (lm *loggingMiddleware) DeletePAT(ctx context.Context, token, patID string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Delete PAT  failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Delete PAT completed successfully", args...)
	}(time.Now())
	return lm.svc.DeletePAT(ctx, token, patID)
}

func (lm *loggingMiddleware) ResetPATSecret(ctx context.Context, token, patID string, duration time.Duration) (pa pat.PAT, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
			slog.String("pat_duration", duration.String()),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Reset PAT secret failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Reset PAT secret completed successfully", args...)
	}(time.Now())
	return lm.svc.ResetPATSecret(ctx, token, patID, duration)
}

func (lm *loggingMiddleware) RevokePATSecret(ctx context.Context, token, patID string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Revoke PAT secret failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Revoke PAT secret completed successfully", args...)
	}(time.Now())
	return lm.svc.RevokePATSecret(ctx, token, patID)
}

func (lm *loggingMiddleware) AddPATScopeEntry(ctx context.Context, token, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (sc pat.Scope, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
			slog.String("platform_entity_type", platformEntityType.String()),
			slog.String("optional_domain_id", optionalDomainID),
			slog.String("optional_domain_entity_type", optionalDomainEntityType.String()),
			slog.String("operation", operation.String()),
			slog.Any("entities", entityIDs),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Add entry to PAT scope failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Add entry to PAT scope completed successfully", args...)
	}(time.Now())
	return lm.svc.AddPATScopeEntry(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (lm *loggingMiddleware) RemovePATScopeEntry(ctx context.Context, token, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (sc pat.Scope, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
			slog.String("platform_entity_type", platformEntityType.String()),
			slog.String("optional_domain_id", optionalDomainID),
			slog.String("optional_domain_entity_type", optionalDomainEntityType.String()),
			slog.String("operation", operation.String()),
			slog.Any("entities", entityIDs),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Remove entry from PAT scope failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Remove entry from PAT scope completed successfully", args...)
	}(time.Now())
	return lm.svc.RemovePATScopeEntry(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (lm *loggingMiddleware) ClearPATAllScopeEntry(ctx context.Context, token, patID string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("pat_id", patID),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Clear all entry from PAT scope failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Clear all entry from PAT scope completed successfully", args...)
	}(time.Now())
	return lm.svc.ClearPATAllScopeEntry(ctx, token, patID)
}

func (lm *loggingMiddleware) IdentifyPAT(ctx context.Context, paToken string) (pa pat.PAT, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Identify PAT failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Identify PAT completed successfully", args...)
	}(time.Now())
	return lm.svc.IdentifyPAT(ctx, paToken)
}

func (lm *loggingMiddleware) AuthorizePAT(ctx context.Context, paToken string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("platform_entity_type", platformEntityType.String()),
			slog.String("optional_domain_id", optionalDomainID),
			slog.String("optional_domain_entity_type", optionalDomainEntityType.String()),
			slog.String("operation", operation.String()),
			slog.Any("entities", entityIDs),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Authorize PAT failed complete successfully", args...)
			return
		}
		lm.logger.Info("Authorize PAT completed successfully", args...)
	}(time.Now())
	return lm.svc.AuthorizePAT(ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (lm *loggingMiddleware) CheckPAT(ctx context.Context, userID, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("user_id", userID),
			slog.String("pat_id", patID),
			slog.String("platform_entity_type", platformEntityType.String()),
			slog.String("optional_domain_id", optionalDomainID),
			slog.String("optional_domain_entity_type", optionalDomainEntityType.String()),
			slog.String("operation", operation.String()),
			slog.Any("entities", entityIDs),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Check PAT failed complete successfully", args...)
			return
		}
		lm.logger.Info("Check PAT completed successfully", args...)
	}(time.Now())
	return lm.svc.CheckPAT(ctx, userID, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}