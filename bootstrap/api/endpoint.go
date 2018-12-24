package api

import (
	"context"

	"github.com/mainflux/mainflux/bootstrap"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/things"
)

func addEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(addReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		config := bootstrap.Config{
			ExternalID:  req.ExternalID,
			ExternalKey: req.ExternalKey,
			MFChannels:  req.Channels,
			Content:     req.Content,
		}

		saved, err := svc.Add(req.key, config)
		if err != nil {
			return nil, err
		}

		res := configRes{
			id:      saved.ID,
			created: true,
		}
		return res, nil
	}
}

func viewEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(entityReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		config, err := svc.View(req.key, req.id)
		if err != nil {
			return nil, err
		}

		res := viewRes{
			ID:          config.ID,
			MFKey:       config.MFKey,
			MFThing:     config.MFThing,
			MFChannels:  config.MFChannels,
			ExternalID:  config.ExternalID,
			ExternalKey: config.ExternalKey,
			Content:     config.Content,
			State:       config.State,
		}
		return res, nil
	}
}

func updateEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(updateReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		config := bootstrap.Config{
			ID:         req.id,
			MFChannels: req.MFChannels,
			Content:    req.Content,
			State:      req.State,
		}

		err := svc.Update(req.key, config)
		if err != nil {
			return nil, err
		}

		res := configRes{
			id:      config.ID,
			created: false,
		}
		return res, nil
	}
}

func listEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(listReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		configs, err := svc.List(req.key, req.filter, req.offset, req.limit)
		if err != nil {
			return nil, err
		}

		res := listRes{
			Configs: []viewRes{},
		}

		for _, cfg := range configs {
			view := viewRes{
				ID:          cfg.ID,
				MFThing:     cfg.MFThing,
				MFKey:       cfg.MFKey,
				MFChannels:  cfg.MFChannels,
				ExternalID:  cfg.ExternalID,
				ExternalKey: cfg.ExternalKey,
				Content:     cfg.Content,
				State:       cfg.State,
			}
			res.Configs = append(res.Configs, view)
		}

		return res, nil
	}
}

func removeEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(entityReq)

		if err := req.validate(); err == things.ErrNotFound {
			return removeRes{}, nil
		}

		if err := svc.Remove(req.key, req.id); err != nil {
			return nil, err
		}

		return removeRes{}, nil
	}
}

func bootstrapEndpoint(svc bootstrap.Service, reader bootstrap.ConfigReader) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(bootstrapReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		cfg, err := svc.Bootstrap(req.key, req.id)
		if err != nil {
			return nil, err
		}

		return reader.ReadConfig(cfg)
	}
}

func stateEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(changeStateReq)

		err := req.validate()
		if err != nil {
			return nil, err
		}

		if err := svc.ChangeState(req.key, req.id, req.State); err != nil {
			return nil, err
		}

		r := stateRes{
			State: req.State,
		}

		return r, nil
	}
}
