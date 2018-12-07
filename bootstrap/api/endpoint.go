package http

import (
	"context"
	"nov/bootstrap"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/things"
)

func addEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(addReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		thing := bootstrap.Thing{
			ExternalID:  req.ExternalID,
			ExternalKey: req.ExternalKey,
			MFChannels:  req.Channels,
			Config:      req.Config,
		}
		saved, err := svc.Add(req.key, thing)
		if err != nil {
			return nil, err
		}

		res := thingRes{
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

		thing, err := svc.View(req.key, req.id)
		if err != nil {
			return nil, err
		}

		res := viewRes{
			ID:         thing.ID,
			Key:        thing.MFKey,
			MFThing:    thing.MFThing,
			MFChannels: thing.MFChannels,
			ExternalID: thing.ExternalID,
			Status:     thing.Status,
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

		thing := bootstrap.Thing{
			ID:         req.id,
			MFChannels: req.MFChannels,
			Config:     req.Config,
			Status:     req.Status,
		}

		err := svc.Update(req.key, thing)
		if err != nil {
			return nil, err
		}

		res := thingRes{
			id:      thing.ID,
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

		things, err := svc.List(req.key, req.offset, req.limit)
		if err != nil {
			return nil, err
		}

		res := listRes{}
		for _, thing := range things {
			view := viewRes{
				ID:         thing.ID,
				MFThing:    thing.MFThing,
				MFChannels: thing.MFChannels,
				ExternalID: thing.ExternalID,
				Status:     thing.Status,
			}
			res.Things = append(res.Things, view)
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
		req := request.(boostrapReq)
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

func statusEndpoint(svc bootstrap.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(changeStatusReq)

		err := req.validate()
		if err != nil {
			return nil, err
		}

		if err := svc.ChangeStatus(req.key, req.id, req.Status); err != nil {
			return nil, err
		}

		r := statusRes{
			Status: req.Status,
		}

		return r, nil
	}
}
