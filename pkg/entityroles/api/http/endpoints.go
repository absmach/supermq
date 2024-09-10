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
		ro, err := svc.AddRole(ctx, req.token, req.entityID, req.RoleName, req.OptionalMembers, req.OptionalMembers)
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
		ros, err := svc.RetrieveAllRoles(ctx, req.token, req.entityID, req.limit, req.offset)
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
		ro, err := svc.RetrieveRole(ctx, req.token, req.entityID, req.roleName)
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
		ro, err := svc.UpdateRoleName(ctx, req.token, req.entityID, req.roleName, req.Name)
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
		if err := svc.RemoveRole(ctx, req.token, req.entityID, req.roleName); err != nil {
			return nil, err
		}
		return deleteRoleRes{}, nil
	}
}
func AddRoleCapabilitiesEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addRoleCapabilitiesReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		caps, err := svc.RoleAddCapabilities(ctx, req.token, req.entityID, req.roleName, req.Capabilities)
		if err != nil {
			return nil, err
		}
		return addRoleCapabilitiesRes{Capabilities: caps}, nil
	}
}
func ListRoleCapabilitiesEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listRoleCapabilitiesReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		caps, err := svc.RoleListCapabilities(ctx, req.token, req.entityID, req.roleName)
		if err != nil {
			return nil, err
		}
		return listRoleCapabilitiesRes{Capabilities: caps}, nil
	}
}
func DeleteRoleCapabilitiesEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteRoleCapabilitiesReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		if err := svc.RoleRemoveCapabilities(ctx, req.token, req.entityID, req.roleName, req.Capabilities); err != nil {
			return nil, err
		}
		return deleteRoleCapabilitiesRes{}, nil
	}
}
func DeleteAllRoleCapabilitiesEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteAllRoleCapabilitiesReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.RoleRemoveAllCapabilities(ctx, req.token, req.entityID, req.roleName); err != nil {
			return nil, err
		}
		return deleteAllRoleCapabilitiesRes{}, nil
	}
}
func AddRoleMembersEndpoint(svc roles.Roles) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addRoleMembersReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}
		members, err := svc.RoleAddMembers(ctx, req.token, req.entityID, req.roleName, req.Members)
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
		mp, err := svc.RoleListMembers(ctx, req.token, req.entityID, req.roleName, req.limit, req.offset)
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
		if err := svc.RoleRemoveMembers(ctx, req.token, req.entityID, req.roleName, req.Members); err != nil {
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
		if err := svc.RoleRemoveAllMembers(ctx, req.token, req.entityID, req.roleName); err != nil {
			return nil, err
		}
		return deleteAllRoleMemberRes{}, nil
	}
}
