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
		ro, err := svc.AddRole(ctx, req.entityID, req.RoleName, roOps, req.OptionalMembers)
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
		ros, err := svc.RetrieveAllRoles(ctx, req.entityID, req.limit, req.offset)
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
		ro, err := svc.RetrieveRole(ctx, req.entityID, req.roleName)
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
		ro, err := svc.UpdateRoleName(ctx, req.entityID, req.roleName, req.Name)
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
		if err := svc.RemoveRole(ctx, req.entityID, req.roleName); err != nil {
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
		ops, err := svc.RoleAddOperation(ctx, req.entityID, req.roleName, roOps)
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
		ops, err := svc.RoleListOperations(ctx, req.entityID, req.roleName)
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
		if err := svc.RoleRemoveOperations(ctx, req.entityID, req.roleName, roOps); err != nil {
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
		if err := svc.RoleRemoveAllOperations(ctx, req.entityID, req.roleName); err != nil {
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
		members, err := svc.RoleAddMembers(ctx, req.entityID, req.roleName, req.Members)
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
		mp, err := svc.RoleListMembers(ctx, req.entityID, req.roleName, req.limit, req.offset)
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
		if err := svc.RoleRemoveMembers(ctx, req.entityID, req.roleName, req.Members); err != nil {
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
		if err := svc.RoleRemoveAllMembers(ctx, req.entityID, req.roleName); err != nil {
			return nil, err
		}
		return deleteAllRoleMemberRes{}, nil
	}
}
