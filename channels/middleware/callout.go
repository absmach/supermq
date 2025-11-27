// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/channels"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rolemw "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
)

var _ channels.Service = (*calloutMiddleware)(nil)

type calloutMiddleware struct {
	svc     channels.Service
	repo    channels.Repository
	callout callout.Callout
	rolemw.RoleManagerCalloutMiddleware
}

func NewCallout(svc channels.Service, repo channels.Repository, callout callout.Callout) (channels.Service, error) {
	call, err := rolemw.NewCallout(policies.ChannelType, svc, callout)
	if err != nil {
		return nil, err
	}

	return &calloutMiddleware{
		svc:                          svc,
		repo:                         repo,
		callout:                      callout,
		RoleManagerCalloutMiddleware: call,
	}, nil
}

func (cm *calloutMiddleware) CreateChannels(ctx context.Context, session authn.Session, chs ...channels.Channel) ([]channels.Channel, []roles.RoleProvision, error) {
	params := map[string]any{
		"entities": chs,
		"count":    len(chs),
	}

	if err := cm.callOut(ctx, session, channels.OpCreateChannel.String(channels.OperationNames), "", params); err != nil {
		return []channels.Channel{}, []roles.RoleProvision{}, err
	}

	return cm.svc.CreateChannels(ctx, session, chs...)
}

func (cm *calloutMiddleware) ViewChannel(ctx context.Context, session authn.Session, id string, withRoles bool) (channels.Channel, error) {
	if err := cm.callOut(ctx, session, channels.OpViewChannel.String(channels.OperationNames), id, nil); err != nil {
		return channels.Channel{}, err
	}

	return cm.svc.ViewChannel(ctx, session, id, withRoles)
}

func (cm *calloutMiddleware) ListChannels(ctx context.Context, session authn.Session, pm channels.Page) (channels.ChannelsPage, error) {
	params := map[string]any{
		"pagemeta": pm,
	}

	if err := cm.callOut(ctx, session, channels.OpListChannels.String(channels.OperationNames), "", params); err != nil {
		return channels.ChannelsPage{}, err
	}

	return cm.svc.ListChannels(ctx, session, pm)
}

func (cm *calloutMiddleware) ListUserChannels(ctx context.Context, session authn.Session, userID string, pm channels.Page) (channels.ChannelsPage, error) {
	params := map[string]any{
		"user_id":  userID,
		"pagemeta": pm,
	}

	if err := cm.callOut(ctx, session, channels.OpListUserChannels.String(channels.OperationNames), "", params); err != nil {
		return channels.ChannelsPage{}, err
	}

	return cm.svc.ListUserChannels(ctx, session, userID, pm)
}

func (cm *calloutMiddleware) UpdateChannel(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if err := cm.callOut(ctx, session, channels.OpUpdateChannel.String(channels.OperationNames), channel.ID, nil); err != nil {
		return channels.Channel{}, err
	}

	return cm.svc.UpdateChannel(ctx, session, channel)
}

func (cm *calloutMiddleware) UpdateChannelTags(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if err := cm.callOut(ctx, session, channels.OpUpdateChannelTags.String(channels.OperationNames), channel.ID, nil); err != nil {
		return channels.Channel{}, err
	}

	return cm.svc.UpdateChannelTags(ctx, session, channel)
}

func (cm *calloutMiddleware) EnableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if err := cm.callOut(ctx, session, channels.OpEnableChannel.String(channels.OperationNames), id, nil); err != nil {
		return channels.Channel{}, err
	}

	return cm.svc.EnableChannel(ctx, session, id)
}

func (cm *calloutMiddleware) DisableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if err := cm.callOut(ctx, session, channels.OpDisableChannel.String(channels.OperationNames), id, nil); err != nil {
		return channels.Channel{}, err
	}

	return cm.svc.DisableChannel(ctx, session, id)
}

func (cm *calloutMiddleware) RemoveChannel(ctx context.Context, session authn.Session, id string) error {
	if err := cm.callOut(ctx, session, channels.OpDeleteChannel.String(channels.OperationNames), id, nil); err != nil {
		return err
	}

	return cm.svc.RemoveChannel(ctx, session, id)
}

func (cm *calloutMiddleware) Connect(ctx context.Context, session authn.Session, chIDs, thIDs []string, connTypes []connections.ConnType) error {
	params := map[string]any{
		"channel_ids":      chIDs,
		"client_ids":       thIDs,
		"connection_types": connTypes,
	}

	if err := cm.callOut(ctx, session, channels.OpConnectClient.String(channels.OperationNames), "", params); err != nil {
		return err
	}

	return cm.svc.Connect(ctx, session, chIDs, thIDs, connTypes)
}

func (cm *calloutMiddleware) Disconnect(ctx context.Context, session authn.Session, chIDs, thIDs []string, connTypes []connections.ConnType) error {
	params := map[string]any{
		"channel_ids":      chIDs,
		"client_ids":       thIDs,
		"connection_types": connTypes,
	}

	if err := cm.callOut(ctx, session, channels.OpDisconnectClient.String(channels.OperationNames), "", params); err != nil {
		return err
	}

	return cm.svc.Disconnect(ctx, session, chIDs, thIDs, connTypes)
}

func (cm *calloutMiddleware) SetParentGroup(ctx context.Context, session authn.Session, parentGroupID string, id string) error {
	params := map[string]any{
		"parent_group_id": parentGroupID,
	}

	if err := cm.callOut(ctx, session, channels.OpSetParentGroup.String(channels.OperationNames), id, params); err != nil {
		return err
	}

	return cm.svc.SetParentGroup(ctx, session, parentGroupID, id)
}

func (cm *calloutMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	ch, err := cm.repo.RetrieveByID(ctx, id)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	if ch.ParentGroup != "" {
		params := map[string]any{
			"parent_group_id": ch.ParentGroup,
		}

		if err := cm.callOut(ctx, session, channels.OpRemoveParentGroup.String(channels.OperationNames), id, params); err != nil {
			return err
		}

		return cm.svc.RemoveParentGroup(ctx, session, id)
	}
	return nil
}

func (cm *calloutMiddleware) callOut(ctx context.Context, session authn.Session, op, entityID string, pld map[string]any) error {
	req := callout.Request{
		BaseRequest: callout.BaseRequest{
			Operation:  op,
			EntityType: policies.ChannelType,
			EntityID:   entityID,
			CallerID:   session.UserID,
			CallerType: policies.UserType,
			DomainID:   session.DomainID,
			Time:       time.Now().UTC(),
		},
		Payload: pld,
	}

	if err := cm.callout.Callout(ctx, req); err != nil {
		return err
	}

	return nil
}
