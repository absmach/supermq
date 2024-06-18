// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/absmach/magistrala/opcua"
)

var _ opcua.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger *slog.Logger
	svc    opcua.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc opcua.Service, logger *slog.Logger) opcua.Service {
	return &loggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}

func (lm loggingMiddleware) CreateThing(ctx context.Context, mgxThing, opcuaNodeID string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("thing_id", mgxThing),
			slog.String("node_id", opcuaNodeID),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Create thing route-map failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Create thing route-map completed successfully", args...)
	}(time.Now())

	return lm.svc.CreateThing(ctx, mgxThing, opcuaNodeID)
}

func (lm loggingMiddleware) UpdateThing(ctx context.Context, mgxThing, opcuaNodeID string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("thing_id", mgxThing),
			slog.String("node_id", opcuaNodeID),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Update thing route-map failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Update thing route-map completed successfully", args...)
	}(time.Now())

	return lm.svc.UpdateThing(ctx, mgxThing, opcuaNodeID)
}

func (lm loggingMiddleware) RemoveThing(ctx context.Context, mgxThing string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("thing_id", mgxThing),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Remove thing route-map failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Remove thing route-map completed successfully", args...)
	}(time.Now())

	return lm.svc.RemoveThing(ctx, mgxThing)
}

func (lm loggingMiddleware) CreateChannel(ctx context.Context, mgxChan, opcuaServerURI string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("channel_id", mgxChan),
			slog.String("server_uri", opcuaServerURI),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Create channel route-map failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Create channel route-map completed successfully", args...)
	}(time.Now())

	return lm.svc.CreateChannel(ctx, mgxChan, opcuaServerURI)
}

func (lm loggingMiddleware) UpdateChannel(ctx context.Context, mgxChanID, opcuaServerURI string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("channel_id", mgxChanID),
			slog.String("server_uri", opcuaServerURI),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Update channel route-map failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Update channel route-map completed successfully", args...)
	}(time.Now())

	return lm.svc.UpdateChannel(ctx, mgxChanID, opcuaServerURI)
}

func (lm loggingMiddleware) RemoveChannel(ctx context.Context, mgxChanID string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("channel_id", mgxChanID),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Remove channel route-map failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Remove channel route-map completed successfully", args...)
	}(time.Now())

	return lm.svc.RemoveChannel(ctx, mgxChanID)
}

func (lm loggingMiddleware) ConnectThing(ctx context.Context, mgxChanID string, mgxThingIDs []string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("channel_id", mgxChanID),
			slog.Any("thing_ids", mgxThingIDs),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Connect thing to channel failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Connect thing to channel completed successfully", args...)
	}(time.Now())

	return lm.svc.ConnectThing(ctx, mgxChanID, mgxThingIDs)
}

func (lm loggingMiddleware) DisconnectThing(ctx context.Context, mgxChanID string, mgxThingIDs []string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("channel_id", mgxChanID),
			slog.Any("thing_ids", mgxThingIDs),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Disconnect thing from channel failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Disconnect thing from channel completed successfully", args...)
	}(time.Now())

	return lm.svc.DisconnectThing(ctx, mgxChanID, mgxThingIDs)
}

func (lm loggingMiddleware) Browse(ctx context.Context, serverURI, namespace, identifier, identifierType string) (nodes []opcua.BrowsedNode, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("server_uri", serverURI),
			slog.String("namespace", namespace),
			slog.String("identifier", identifier),
			slog.String("identifier_type", identifierType),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.WarnContext(ctx, "Browse available nodes failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Browse available nodes completed successfully", args...)
	}(time.Now())

	return lm.svc.Browse(ctx, serverURI, namespace, identifier, identifierType)
}
