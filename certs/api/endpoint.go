// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/certs"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/go-kit/kit/endpoint"
)

func issueCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(addCertsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		res, err := svc.IssueCert(ctx, req.domainID, req.token, req.ClientID, req.TTL)
		if err != nil {
			return certsRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}

		return certsRes{
			SerialNumber: res.SerialNumber,
			ClientID:     res.ClientID,
			Certificate:  res.Certificate,
			Key:          res.Key,
			CAChain:      res.CAChain,
			IssuingCA:    res.IssuingCA,
			ExpiryTime:   res.ExpiryTime,
			Revoked:      res.Revoked,
			issued:       true,
		}, nil
	}
}

func listSerials(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(listReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		page, err := svc.ListSerials(ctx, req.clientID, req.pm)
		if err != nil {
			return certsPageRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		res := certsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Certs: []certsRes{},
		}

		for _, cert := range page.Certificates {
			cr := certsRes{
				SerialNumber: cert.SerialNumber,
				ExpiryTime:   cert.ExpiryTime,
				ClientID:     cert.ClientID,
				Revoked:      cert.Revoked,
			}
			res.Certs = append(res.Certs, cr)
		}
		return res, nil
	}
}

func viewCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(viewReq)
		if err := req.validate(); err != nil {
			return certsRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}

		cert, err := svc.ViewCert(ctx, req.serialID)
		if err != nil {
			return certsRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}

		return certsRes{
			ClientID:     cert.ClientID,
			Certificate:  cert.Certificate,
			Key:          cert.Key,
			SerialNumber: cert.SerialNumber,
			ExpiryTime:   cert.ExpiryTime,
			Revoked:      cert.Revoked,
			issued:       false,
		}, nil
	}
}

func revokeAllCerts(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(revokeAllReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		res, err := svc.RevokeCert(ctx, req.domainID, req.token, req.clientID)
		if err != nil {
			return nil, err
		}
		return revokeCertsRes{
			RevocationTime: res.RevocationTime,
		}, nil
	}
}

func revokeBySerial(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(revokeBySerialReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		res, err := svc.RevokeBySerial(ctx, req.serialID)
		if err != nil {
			return nil, err
		}
		return revokeCertsRes{
			RevocationTime: res.RevocationTime,
		}, nil
	}
}
