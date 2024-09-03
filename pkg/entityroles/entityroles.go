package entityroles

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/roles"
	"github.com/teris-io/shortid"
)

var (
	errRollbackRoles    = errors.New("failed to rollback roles")
	errRollbackPolicies = errors.New("failed to rollback roles")
	errAddPolicies      = errors.New("failed to add policies")
)
var _ roles.Roles = (*rolesSvc)(nil)

type rolesSvc struct {
	entityType   string
	db           roles.Repository
	auth         magistrala.AuthServiceClient
	operations   []roles.Operation
	builtInRoles map[string][]roles.Operation
}

func NewRole(entityType string) roles.Roles {
	return &rolesSvc{entityType: entityType}
}

func (r *rolesSvc) isOperationAllowed(op roles.Operation) bool {
	for _, aop := range r.operations {
		if aop == op {
			return true
		}
	}
	return false
}

func (r *rolesSvc) validateOperations(operations []roles.Operation) error {
	for _, op := range operations {
		roOp := roles.Operation(op)
		if !r.isOperationAllowed(roOp) {
			return errors.Wrap(svcerr.ErrMalformedEntity, fmt.Errorf("invalid operation %s ", op))
		}
	}
	return nil
}

func (r *rolesSvc) AddNewEntityRoles(ctx context.Context, entityID string, newEntityDefaultRoleMembers map[string][]string, optionalPolicies []roles.OptionalPolicy) (ros []roles.Role, err error) {
	var newRoleProvisions []roles.RoleProvision
	prs := []*magistrala.AddPolicyReq{}

	for defaultRole, defaultRoleMembers := range newEntityDefaultRoleMembers {
		operations, ok := r.builtInRoles[defaultRole]
		if !ok {
			return []roles.Role{}, fmt.Errorf("default role %s not found in in-built roles", defaultRole)
		}

		// There an option to have id as entityID_roleName where in roleName all space are removed with _ and starts with letter and supports only alphanumeric, space and hyphen
		id, err := shortid.Generate()

		if err != nil {
			return []roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
		}

		if err := r.validateOperations(operations); err != nil {
			return []roles.Role{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
		}

		newRoleProvisions = append(newRoleProvisions, roles.RoleProvision{
			Role: roles.Role{
				ID:        id,
				Name:      defaultRole,
				EntityID:  entityID,
				CreatedAt: time.Now(),
				// CreatedBy: username,
			},
			OptionalOperations: operations,
			OptionalMembers:    defaultRoleMembers,
		})

		for _, op := range operations {
			prs = append(prs, &magistrala.AddPolicyReq{
				SubjectType:     "role",
				SubjectRelation: "member",
				Subject:         id,
				Relation:        string(op),
				Object:          entityID,
				ObjectType:      r.entityType,
			})
		}

		for _, member := range defaultRoleMembers {
			prs = append(prs, &magistrala.AddPolicyReq{
				SubjectType: "user",
				Subject:     member,
				Relation:    "member",
				Object:      id,
				ObjectType:  "role",
			})
		}

	}

	if len(optionalPolicies) > 0 {
		for _, policy := range optionalPolicies {
			prs = append(prs, &magistrala.AddPolicyReq{
				Domain:          policy.Namespace,
				SubjectType:     policy.SubjectType,
				SubjectRelation: policy.SubjectRelation,
				Subject:         policy.Subject,
				Relation:        policy.Relation,
				Permission:      policy.Permission,
				Object:          policy.Object,
				ObjectType:      policy.ObjectType,
			})
		}
	}

	resp, err := r.auth.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})

	if err != nil {
		return []roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	if !resp.Added {
		return []roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, errAddPolicies)
	}
	defer func() {
		if err != nil {
			if errRollBack := r.AddPolicyRollback(ctx, prs); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRoles, errRollBack))
			}
		}
	}()

	newRoles, err := r.db.Add(ctx, newRoleProvisions)
	if err != nil {
		return []roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	return newRoles, nil
}

func (r *rolesSvc) AddPolicyRollback(ctx context.Context, prs []*magistrala.AddPolicyReq) error {
	delPrs := []*magistrala.DeletePolicyReq{}

	for _, pr := range prs {
		delPrs = append(delPrs, &magistrala.DeletePolicyReq{
			Domain:          pr.Domain,
			SubjectType:     pr.SubjectType,
			SubjectRelation: pr.SubjectRelation,
			Subject:         pr.Subject,
			Relation:        pr.Relation,
			Permission:      pr.Permission,
			Object:          pr.Object,
			ObjectType:      pr.ObjectType,
		})
	}
	resp, err := r.auth.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: delPrs})
	if err != nil {
		return errors.Wrap(errRollbackPolicies, err)
	}
	if !resp.Deleted {
		return errRollbackPolicies
	}
	return nil
}
func (r *rolesSvc) Add(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	// There an option to have id as entityID_roleName where in roleName all space are removed with _ and starts with letter and supports only alphanumeric, space and hyphen
	id, err := shortid.Generate()
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	if err := r.validateOperations(optionalOperations); err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
	}

	newRoleProvisions := []roles.RoleProvision{
		{
			Role: roles.Role{
				ID:        id,
				Name:      roleName,
				EntityID:  entityID,
				CreatedAt: time.Now(),
				// CreatedBy: username,
			},
			OptionalOperations: optionalOperations,
			OptionalMembers:    optionalMembers,
		},
	}
	prs := []*magistrala.AddPolicyReq{}

	for _, op := range optionalOperations {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType:     "role",
			SubjectRelation: "member",
			Subject:         id,
			Relation:        string(op),
			Object:          entityID,
			ObjectType:      r.entityType,
		})
	}

	for _, member := range optionalMembers {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType:     "user",
			SubjectRelation: member,
			Subject:         id,
			Relation:        "member",
			Object:          id,
			ObjectType:      "role",
		})
	}

	resp, err := r.auth.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})

	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	if !resp.Added {
		return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, errAddPolicies)
	}
	defer func() {
		if err != nil {
			if errRollBack := r.AddPolicyRollback(ctx, prs); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRoles, errRollBack))
			}
		}
	}()

	newRoles, err := r.db.Add(ctx, newRoleProvisions)
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	if len(newRoles) == 0 {
		return roles.Role{}, svcerr.ErrCreateEntity
	}

	return newRoles[1], nil
}

func (r *rolesSvc) Remove(ctx context.Context, entityID, roleName string) error {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	req := &magistrala.DeletePolicyFilterReq{
		SubjectType: "role",
		Subject:     ro.ID,
	}
	resp, err := r.auth.DeletePolicyFilter(ctx, req)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	if !resp.Deleted {
		return svcerr.ErrRemoveEntity
	}

	if err := r.db.Remove(ctx, []string{ro.ID}); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r *rolesSvc) UpdateName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, oldRoleName)
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	ro, err = r.db.Update(ctx, roles.Role{
		ID:       ro.ID,
		EntityID: entityID,
		Name:     newRoleName,
	})
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return ro, nil
}

func (r *rolesSvc) Retrieve(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return ro, nil
}

func (r *rolesSvc) RetrieveAll(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	ros, err := r.db.RetrieveAll(ctx, entityID, limit, offset)
	if err != nil {
		return roles.RolePage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return ros, nil
}

func (r *rolesSvc) AddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return []roles.Operation{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	if len(operations) == 0 {
		return []roles.Operation{}, svcerr.ErrMalformedEntity
	}

	if err := r.validateOperations(operations); err != nil {
		return []roles.Operation{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
	}

	prs := []*magistrala.AddPolicyReq{}
	for _, op := range operations {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType:     "role",
			SubjectRelation: "member",
			Subject:         ro.ID,
			Relation:        string(op),
			Object:          entityID,
			ObjectType:      r.entityType,
		})
	}

	resp, err := r.auth.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})
	if err != nil {
		return []roles.Operation{}, errors.Wrap(svcerr.ErrAddPolicies, err)
	}
	if !resp.Added {
		return []roles.Operation{}, svcerr.ErrAddPolicies
	}

	defer func() {
		if err != nil {
			if errRollBack := r.AddPolicyRollback(ctx, prs); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRoles, errRollBack))
			}
		}
	}()

	ro.UpdatedAt = time.Now()
	// ro.UpdateBy = userID

	resOps, err := r.db.AddOperation(ctx, ro, operations)
	if err != nil {
		return []roles.Operation{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	return resOps, nil
}

func (r *rolesSvc) ListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return []roles.Operation{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	ops, err := r.db.ListOperations(ctx, ro.ID)
	if err != nil {
		return []roles.Operation{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return ops, nil

}

func (r *rolesSvc) CheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return false, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	result, err := r.db.CheckOperationsExists(ctx, ro.ID, operations)
	if err != nil {
		return true, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return result, nil
}

func (r *rolesSvc) RemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	if len(operations) == 0 {
		return svcerr.ErrMalformedEntity
	}

	prs := []*magistrala.DeletePolicyReq{}
	for _, op := range operations {
		prs = append(prs, &magistrala.DeletePolicyReq{
			SubjectType:     "role",
			SubjectRelation: "member",
			Subject:         ro.ID,
			Relation:        string(op),
			Object:          entityID,
			ObjectType:      r.entityType,
		})
	}

	resp, err := r.auth.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: prs})
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	if !resp.Deleted {
		return svcerr.ErrDeletePolicies
	}

	ro.UpdatedAt = time.Now()
	// ro.UpdatedBy = userID
	if err := r.db.RemoveOperations(ctx, ro, operations); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r *rolesSvc) RemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	prs := &magistrala.DeletePolicyFilterReq{
		SubjectType: "role",
		Subject:     ro.ID,
	}

	resp, err := r.auth.DeletePolicyFilter(ctx, prs)
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	if !resp.Deleted {
		return svcerr.ErrDeletePolicies
	}

	ro.UpdatedAt = time.Now()
	// ro.UpdatedBy = userID

	if err := r.db.RemoveAllOperations(ctx, ro); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r *rolesSvc) AddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	if len(members) == 0 {
		return []string{}, svcerr.ErrMalformedEntity
	}

	prs := []*magistrala.AddPolicyReq{}
	for _, mem := range members {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType: "user",
			Subject:     mem,
			Relation:    "member",
			Object:      ro.ID,
			ObjectType:  "role",
		})
	}

	resp, err := r.auth.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrAddPolicies, err)
	}
	if !resp.Added {
		return []string{}, svcerr.ErrAddPolicies
	}

	defer func() {
		if err != nil {
			if errRollBack := r.AddPolicyRollback(ctx, prs); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRoles, errRollBack))
			}
		}
	}()

	ro.UpdatedAt = time.Now()
	// ro.UpdateBy = userID

	mems, err := r.db.AddMembers(ctx, ro, members)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	return mems, nil
}

func (r *rolesSvc) ListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return roles.MembersPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	mp, err := r.db.ListMembers(ctx, ro.ID, limit, offset)
	if err != nil {
		return roles.MembersPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return mp, nil
}

func (r *rolesSvc) CheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return false, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	result, err := r.db.CheckMembersExists(ctx, ro.ID, members)
	if err != nil {
		return true, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return result, nil
}

func (r *rolesSvc) RemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	if len(members) == 0 {
		return svcerr.ErrMalformedEntity
	}

	prs := []*magistrala.DeletePolicyReq{}
	for _, mem := range members {
		prs = append(prs, &magistrala.DeletePolicyReq{
			SubjectType: "user",
			Subject:     mem,
			Relation:    "member",
			Object:      ro.ID,
			ObjectType:  "role",
		})
	}

	resp, err := r.auth.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: prs})
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	if !resp.Deleted {
		return svcerr.ErrDeletePolicies
	}

	ro.UpdatedAt = time.Now()
	// ro.UpdatedBy = userID
	if err := r.db.RemoveMembers(ctx, ro, members); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r *rolesSvc) RemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	ro, err := r.db.RetrieveByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	prs := &magistrala.DeletePolicyFilterReq{
		ObjectType:  "role",
		Object:      ro.ID,
		SubjectType: "user",
	}

	resp, err := r.auth.DeletePolicyFilter(ctx, prs)
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	if !resp.Deleted {
		return svcerr.ErrDeletePolicies
	}

	ro.UpdatedAt = time.Now()
	// ro.UpdatedBy = userID

	if err := r.db.RemoveAllMembers(ctx, ro); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}
