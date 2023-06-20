// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	notifiers "github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

func createSubscriptionEndpoint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createSubReq)
		if err := req.validate(); err != nil {
			return createSubRes{}, errors.Wrap(err, apiutil.ErrValidation)
		}
		sub := notifiers.Subscription{
			Contact: req.Contact,
			Topic:   req.Topic,
		}
		id, err := svc.CreateSubscription(ctx, req.token, sub)
		if err != nil {
			return createSubRes{}, errors.Wrap(err, apiutil.ErrValidation)
		}
		ucr := createSubRes{
			ID: id,
		}

		return ucr, nil
	}
}

func viewSubscriptionEndpint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(subReq)
		if err := req.validate(); err != nil {
			return viewSubRes{}, errors.Wrap(err, apiutil.ErrValidation)
		}
		sub, err := svc.ViewSubscription(ctx, req.token, req.id)
		if err != nil {
			return viewSubRes{}, errors.Wrap(err, apiutil.ErrValidation)
		}
		res := viewSubRes{
			ID:      sub.ID,
			OwnerID: sub.OwnerID,
			Contact: sub.Contact,
			Topic:   sub.Topic,
		}
		return res, nil
	}
}

func listSubscriptionsEndpoint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listSubsReq)
		if err := req.validate(); err != nil {
			return listSubsRes{}, errors.Wrap(err, apiutil.ErrValidation)
		}
		pm := notifiers.PageMetadata{
			Topic:   req.topic,
			Contact: req.contact,
			Offset:  req.offset,
			Limit:   int(req.limit),
		}
		page, err := svc.ListSubscriptions(ctx, req.token, pm)
		if err != nil {
			return listSubsRes{}, errors.Wrap(err, apiutil.ErrValidation)
		}
		res := listSubsRes{
			Offset: page.Offset,
			Limit:  page.Limit,
			Total:  page.Total,
		}
		for _, sub := range page.Subscriptions {
			r := viewSubRes{
				ID:      sub.ID,
				OwnerID: sub.OwnerID,
				Contact: sub.Contact,
				Topic:   sub.Topic,
			}
			res.Subscriptions = append(res.Subscriptions, r)
		}

		return res, nil
	}
}

func deleteSubscriptionEndpint(svc notifiers.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(subReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(err, apiutil.ErrValidation)
		}
		if err := svc.RemoveSubscription(ctx, req.token, req.id); err != nil {
			return nil, errors.Wrap(err, apiutil.ErrValidation)
		}
		return removeSubRes{}, nil
	}
}
