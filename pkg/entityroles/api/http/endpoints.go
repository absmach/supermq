package http

import (
	"context"

	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/absmach/magistrala/pkg/roles"
	"github.com/go-kit/kit/endpoint"
)

func CreateRoleEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createRoleReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		roOps := []roles.Operation{}
		for _, ops := range req.OptionalOperations {
			roOps = append(roOps, roles.Operation(ops))
		}
		ro, err := svc.Add(ctx, req.EntityID, req.RoleName, roOps, req.OptionalMembers)
		if err != nil {
			return nil, err
		}
		return createRoleRes{Role: ro}, nil
	}
}

func ListRolesEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listRolesReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		ros, err := svc.RetrieveAll(ctx, req.entityID, req.limit, req.offset)
		if err != nil {
			return nil, err
		}
		return listRolesRes{RolePage: ros}, nil
	}
}

func ViewRoleEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewRoleReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		ro, err := svc.Retrieve(ctx, req.entityID, req.roleName)
		if err != nil {
			return nil, err
		}
		return viewRoleRes{Role: ro}, nil
	}
}

func UpdateRoleEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateRoleReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		ro, err := svc.UpdateName(ctx, req.entityID, req.roleName, req.Name)
		if err != nil {
			return nil, err
		}
		return updateRoleRes{Role: ro}, nil
	}
}
func DeleteRoleEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteRoleReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.Remove(ctx, req.entityID, req.roleName); err != nil {
			return nil, err
		}
		return deleteRoleRes{}, nil
	}
}
func AddRoleOperationsEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addRoleOperationsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		roOps := []roles.Operation{}
		for _, ops := range req.Operations {
			roOps = append(roOps, roles.Operation(ops))
		}
		ops, err := svc.AddOperation(ctx, req.entityID, req.roleName, roOps)
		if err != nil {
			return nil, err
		}
		return addRoleOperationsRes{Operations: ops}, nil
	}
}
func ListRoleOperationsEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listRoleOperationsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		ops, err := svc.ListOperations(ctx, req.entityID, req.roleName)
		if err != nil {
			return nil, err
		}
		return listRoleOperationsRes{Operations: ops}, nil
	}
}
func DeleteRoleOperationsEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteRoleOperationsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		roOps := []roles.Operation{}
		for _, ops := range req.Operations {
			roOps = append(roOps, roles.Operation(ops))
		}
		if err := svc.RemoveOperations(ctx, req.entityID, req.roleName, roOps); err != nil {
			return nil, err
		}
		return deleteRoleOperationsRes{}, nil
	}
}
func DeleteAllRoleOperationsEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteAllRoleOperationsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.RemoveAllOperations(ctx, req.entityID, req.roleName); err != nil {
			return nil, err
		}
		return deleteAllRoleOperationsRes{}, nil
	}
}
func AddRoleMembersEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addRoleMembersReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		members, err := svc.AddMembers(ctx, req.entityID, req.roleName, req.Members)
		if err != nil {
			return nil, err
		}
		return addRoleMembersRes{members}, nil
	}
}
func ListRoleMembersEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listRoleMembersReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		mp, err := svc.ListMembers(ctx, req.entityID, req.roleName, req.limit, req.offset)
		if err != nil {
			return nil, err
		}
		return listRoleMembersRes{mp}, nil
	}
}
func DeleteRoleMembersEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteRoleMembersReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.RemoveMembers(ctx, req.entityID, req.roleName, req.Members); err != nil {
			return nil, err
		}
		return deleteRoleMembersRes{}, nil
	}
}
func DeleteAllRoleMembersEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteAllRoleMembersReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.RemoveAllMembers(ctx, req.entityID, req.roleName); err != nil {
			return nil, err
		}
		return deleteAllRoleMemberRes{}, nil
	}
}
