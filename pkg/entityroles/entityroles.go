package entityroles

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/roles"
)

type identity struct {
	ID       string
	DomainID string
	UserID   string
}

type EntityRoleOperationKey int

const (
	OpAddRole EntityRoleOperationKey = iota
	OpRemoveRole
	OpUpdateRoleName
	OpRetrieveRole
	OpRetrieveAllRoles
	OpRoleAddOperations
	OpRoleListOperations
	OpRoleCheckOperationsExists
	OpRoleRemoveOperations
	OpRoleRemoveAllOperations
	OpRoleAddMembers
	OpRoleListMembers
	OpRoleCheckMembersExists
	OpRoleRemoveMembers
	OpRoleRemoveAllMembers
)

func (erok EntityRoleOperationKey) String() string {
	names := [...]string{
		"OpAddRole",
		"OpRemoveRole",
		"OpUpdateRoleName",
		"OpRetrieveRole",
		"OpRetrieveAllRoles",
		"OpRoleAddOperations",
		"OpRoleListOperations",
		"OpRoleCheckOperationsExists",
		"OpRoleRemoveOperations",
		"OpRoleRemoveAllOperations",
		"OpRoleAddMembers",
		"OpRoleListMembers",
		"OpRoleCheckMembersExists",
		"OpRoleRemoveMembers",
		"OpRoleRemoveAllMembers",
	}

	if int(erok) < 0 || int(erok) >= len(names) {
		return fmt.Sprintf("UnknownEntityRoleOperationKey(%d)", erok)
	}

	return names[erok]
}

var expectedEntityRoleOperationKey = []EntityRoleOperationKey{
	OpAddRole,
	OpRemoveRole,
	OpUpdateRoleName,
	OpRetrieveRole,
	OpRetrieveAllRoles,
	OpRoleAddOperations,
	OpRoleListOperations,
	OpRoleCheckOperationsExists,
	OpRoleRemoveOperations,
	OpRoleRemoveAllOperations,
	OpRoleAddMembers,
	OpRoleListMembers,
	OpRoleCheckMembersExists,
	OpRoleRemoveMembers,
	OpRoleRemoveAllMembers,
}

type EntityRolesOperations map[EntityRoleOperationKey]roles.Operation

func NewEntityRolesOperations(newEro map[EntityRoleOperationKey]roles.Operation) (EntityRolesOperations, error) {
	ero := EntityRolesOperations(newEro)
	if err := ero.Validate(); err != nil {
		return EntityRolesOperations{}, err
	}
	return ero, nil
}

func (ero EntityRolesOperations) isKeyRequired(erok EntityRoleOperationKey) bool {
	for _, key := range expectedEntityRoleOperationKey {
		if key == erok {
			return true
		}
	}
	return false
}

func (ero EntityRolesOperations) Add(erok EntityRoleOperationKey, op roles.Operation) error {
	if !ero.isKeyRequired(erok) {
		return fmt.Errorf("%v is not a valid role operation key", erok)
	}
	ero[erok] = op
	return nil
}

func (ero EntityRolesOperations) Validate() error {
	for erok := range ero {
		if !ero.isKeyRequired(erok) {
			return fmt.Errorf("EntityRolesOperations: \"%s\" is not a valid entity roles operation key", erok.String())
		}
	}
	for _, erok := range expectedEntityRoleOperationKey {
		if _, ok := ero[erok]; !ok {
			return fmt.Errorf("EntityRolesOperations: \"%s\" is not provided", erok.String())
		}
	}
	return nil
}

var (
	errRollbackRoles    = errors.New("failed to rollback roles")
	errRollbackPolicies = errors.New("failed to rollback roles")
	errAddPolicies      = errors.New("failed to add policies")
)
var _ roles.Roles = (*RolesSvc)(nil)

type RolesSvc struct {
	entityType   string
	repo         roles.Repository
	idProvider   magistrala.IDProvider
	auth         magistrala.AuthServiceClient
	operations   []roles.Operation
	builtInRoles map[roles.BuiltInRoleName][]roles.Operation
	ero          EntityRolesOperations
}

func NewRolesSvc(entityType string, repo roles.Repository, idProvider magistrala.IDProvider, auth magistrala.AuthServiceClient, operations []roles.Operation, builtInRoles map[roles.BuiltInRoleName][]roles.Operation, ero EntityRolesOperations) (RolesSvc, error) {
	rolesSvc := RolesSvc{
		entityType:   entityType,
		repo:         repo,
		idProvider:   idProvider,
		auth:         auth,
		operations:   operations,
		builtInRoles: builtInRoles,
		ero:          ero,
	}
	if err := rolesSvc.validate(); err != nil {
		return RolesSvc{}, err
	}
	return rolesSvc, nil
}

func (r RolesSvc) validate() error {
	for erok, op := range r.ero {
		if !r.isOperationAllowed(op) {
			return fmt.Errorf("Invalid EntityRolesOperations: entityroles operation \"%s\" have operation \"%s\" which is not present in entity operations", erok.String(), op)
		}
	}
	return nil
}

func (r RolesSvc) isOperationAllowed(op roles.Operation) bool {
	for _, aop := range r.operations {
		if aop == op {
			return true
		}
	}
	return false
}

func toRolesOperations(ops []string) []roles.Operation {
	roOps := []roles.Operation{}
	for _, op := range ops {
		roOps = append(roOps, roles.Operation(op))
	}
	return roOps
}

func roleOperationsToString(roOps []roles.Operation) []string {
	ops := []string{}
	for _, roOp := range roOps {
		ops = append(ops, roOp.String())
	}
	return ops
}

func roleMembersToString(roMems []roles.Member) []string {
	mems := []string{}
	for _, roMem := range roMems {
		mems = append(mems, roMem.String())
	}
	return mems
}
func (r RolesSvc) validateOperations(operations []roles.Operation) error {
	for _, op := range operations {
		roOp := roles.Operation(op)
		if !r.isOperationAllowed(roOp) {
			return errors.Wrap(svcerr.ErrMalformedEntity, fmt.Errorf("invalid operation %s ", op))
		}
	}
	return nil
}

func (r RolesSvc) AddRole(ctx context.Context, token, entityID string, roleName string, optionalOperations []string, optionalMembers []string) (fnRole roles.Role, fnErr error) {

	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return roles.Role{}, err
	}

	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpAddRole].String(),
	}); err != nil {
		return roles.Role{}, err
	}

	// ToDo: Research: Discuss: There an option to have id as entityID_roleName where in roleName all space are removed with _ and starts with letter and supports only alphanumeric, space and hyphen
	id, err := r.idProvider.ID()
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	if err := r.validateOperations(toRolesOperations(optionalOperations)); err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
	}

	newRoleProvisions := []roles.RoleProvision{
		{
			Role: roles.Role{
				ID:        id,
				Name:      roleName,
				EntityID:  entityID,
				CreatedAt: time.Now(),
				CreatedBy: userInfo.UserID,
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
			Relation:        op,
			Object:          entityID,
			ObjectType:      r.entityType,
		})
	}

	for _, member := range optionalMembers {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType: "user",
			Subject:     auth.EncodeDomainUserID(userInfo.DomainID, member),
			Relation:    "member",
			Object:      id,
			ObjectType:  "role",
		})
	}

	if len(prs) > 0 {
		resp, err := r.auth.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})

		if err != nil {
			return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
		}
		if !resp.Added {
			return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, errAddPolicies)
		}
		defer func() {
			if fnErr != nil {
				if errRollBack := r.addPolicyRollback(ctx, prs); errRollBack != nil {
					fnErr = errors.Wrap(fnErr, errors.Wrap(errRollbackRoles, errRollBack))
				}
			}
		}()
	}

	newRoles, err := r.repo.AddRoles(context.Background(), newRoleProvisions)
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	if len(newRoles) == 0 {
		return roles.Role{}, svcerr.ErrCreateEntity
	}

	return newRoles[0], nil
}

func (r RolesSvc) RemoveRole(ctx context.Context, token, entityID, roleName string) error {

	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return err
	}

	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRemoveRole].String(),
	}); err != nil {
		return err
	}

	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
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

	// ToDo: Add Role as Object in DeletePolicyFilter

	if err := r.repo.RemoveRoles(ctx, []string{ro.ID}); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r RolesSvc) UpdateRoleName(ctx context.Context, token, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return roles.Role{}, err
	}

	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpUpdateRoleName].String(),
	}); err != nil {
		return roles.Role{}, err
	}

	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, oldRoleName)
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	ro, err = r.repo.UpdateRole(ctx, roles.Role{
		ID:        ro.ID,
		EntityID:  entityID,
		Name:      newRoleName,
		UpdatedBy: userInfo.UserID,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return ro, nil
}

func (r RolesSvc) RetrieveRole(ctx context.Context, token, entityID, roleName string) (roles.Role, error) {
	userInfo, err := r.identify(ctx, token)
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRetrieveRole].String(),
	}); err != nil {
		return roles.Role{}, err
	}

	if err != nil {
		return roles.Role{}, err
	}

	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return ro, nil
}

func (r RolesSvc) RetrieveAllRoles(ctx context.Context, token, entityID string, limit, offset uint64) (roles.RolePage, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return roles.RolePage{}, err
	}

	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRetrieveAllRoles].String(),
	}); err != nil {
		return roles.RolePage{}, err
	}

	ros, err := r.repo.RetrieveAllRoles(ctx, entityID, limit, offset)
	if err != nil {
		return roles.RolePage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return ros, nil
}

func (r RolesSvc) RoleAddOperations(ctx context.Context, token, entityID, roleName string, operations []string) (fnOps []string, fnErr error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return []string{}, err
	}

	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleAddOperations].String(),
	}); err != nil {
		return []string{}, err
	}

	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	if len(operations) == 0 {
		return []string{}, svcerr.ErrMalformedEntity
	}

	if err := r.validateOperations(toRolesOperations(operations)); err != nil {
		return []string{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
	}

	prs := []*magistrala.AddPolicyReq{}
	for _, op := range operations {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType:     "role",
			SubjectRelation: "member",
			Subject:         ro.ID,
			Relation:        op,
			Object:          entityID,
			ObjectType:      r.entityType,
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
		if fnErr != nil {
			if errRollBack := r.addPolicyRollback(ctx, prs); errRollBack != nil {
				fnErr = errors.Wrap(fnErr, errors.Wrap(errRollbackRoles, errRollBack))
			}
		}
	}()

	ro.UpdatedAt = time.Now()
	ro.UpdatedBy = userInfo.UserID

	resOps, err := r.repo.RoleAddOperation(ctx, ro, operations)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	return resOps, nil
}

func (r RolesSvc) RoleListOperations(ctx context.Context, token, entityID, roleName string) ([]string, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return []string{}, err
	}

	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleListOperations].String(),
	}); err != nil {
		return []string{}, err
	}

	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	ops, err := r.repo.RoleListOperations(ctx, ro.ID)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return ops, nil

}

func (r RolesSvc) RoleCheckOperationsExists(ctx context.Context, token, entityID, roleName string, operations []string) (bool, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return false, err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleCheckOperationsExists].String(),
	}); err != nil {
		return false, err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return false, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	result, err := r.repo.RoleCheckOperationsExists(ctx, ro.ID, operations)
	if err != nil {
		return true, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return result, nil
}

func (r RolesSvc) RoleRemoveOperations(ctx context.Context, token, entityID, roleName string, operations []string) (err error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleRemoveOperations].String(),
	}); err != nil {
		return err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
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
			Relation:        op,
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
	ro.UpdatedBy = userInfo.UserID
	if err := r.repo.RoleRemoveOperations(ctx, ro, operations); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r RolesSvc) RoleRemoveAllOperations(ctx context.Context, token, entityID, roleName string) error {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleRemoveAllOperations].String(),
	}); err != nil {
		return err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
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
	ro.UpdatedBy = userInfo.UserID

	if err := r.repo.RoleRemoveAllOperations(ctx, ro); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r RolesSvc) RoleAddMembers(ctx context.Context, token, entityID, roleName string, members []string) (fnMems []string, fnErr error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return []string{}, err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleAddMembers].String(),
	}); err != nil {
		return []string{}, err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
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
			Subject:     auth.EncodeDomainUserID(userInfo.DomainID, mem),
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
		if fnErr != nil {
			if errRollBack := r.addPolicyRollback(ctx, prs); errRollBack != nil {
				fnErr = errors.Wrap(fnErr, errors.Wrap(errRollbackRoles, errRollBack))
			}
		}
	}()

	ro.UpdatedAt = time.Now()
	ro.UpdatedBy = userInfo.UserID

	mems, err := r.repo.RoleAddMembers(ctx, ro, members)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	return mems, nil
}

func (r RolesSvc) RoleListMembers(ctx context.Context, token, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return roles.MembersPage{}, err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleListMembers].String(),
	}); err != nil {
		return roles.MembersPage{}, err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return roles.MembersPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	mp, err := r.repo.RoleListMembers(ctx, ro.ID, limit, offset)
	if err != nil {
		return roles.MembersPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return mp, nil
}

func (r RolesSvc) RoleCheckMembersExists(ctx context.Context, token, entityID, roleName string, members []string) (bool, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return false, err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleCheckMembersExists].String(),
	}); err != nil {
		return false, err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return false, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	result, err := r.repo.RoleCheckMembersExists(ctx, ro.ID, members)
	if err != nil {
		return true, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return result, nil
}

func (r RolesSvc) RoleRemoveMembers(ctx context.Context, token, entityID, roleName string, members []string) (err error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleRemoveMembers].String(),
	}); err != nil {
		return err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
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
			Subject:     auth.EncodeDomainUserID(userInfo.DomainID, mem),
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
	if err := r.repo.RoleRemoveMembers(ctx, ro, members); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r RolesSvc) RoleRemoveAllMembers(ctx context.Context, token, entityID, roleName string) (err error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := r.authorize(ctx, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
		Permission:  r.ero[OpRoleRemoveAllMembers].String(),
	}); err != nil {
		return err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
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
	ro.UpdatedBy = userInfo.UserID

	if err := r.repo.RoleRemoveAllMembers(ctx, ro); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r RolesSvc) AddNewEntityRoles(ctx context.Context, userID, domainID, entityID string, newBuiltInRoleMembers map[roles.BuiltInRoleName][]roles.Member, optionalPolicies []roles.OptionalPolicy) (fnRolesProvision []roles.RoleProvision, fnErr error) {
	var newRolesProvision []roles.RoleProvision
	prs := []*magistrala.AddPolicyReq{}

	for defaultRole, defaultRoleMembers := range newBuiltInRoleMembers {
		operations, ok := r.builtInRoles[defaultRole]
		if !ok {
			return []roles.RoleProvision{}, fmt.Errorf("default role %s not found in in-built roles", defaultRole)
		}

		// There an option to have id as entityID_roleName where in roleName all space are removed with _ and starts with letter and supports only alphanumeric, space and hyphen
		id, err := r.idProvider.ID()
		if err != nil {
			return []roles.RoleProvision{}, errors.Wrap(svcerr.ErrCreateEntity, err)
		}

		if err := r.validateOperations(operations); err != nil {
			return []roles.RoleProvision{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
		}

		members := roleMembersToString(defaultRoleMembers)
		ops := roleOperationsToString(operations)

		newRolesProvision = append(newRolesProvision, roles.RoleProvision{
			Role: roles.Role{
				ID:        id,
				Name:      defaultRole.String(),
				EntityID:  entityID,
				CreatedAt: time.Now(),
				CreatedBy: userID,
			},
			OptionalOperations: ops,
			OptionalMembers:    members,
		})

		for _, op := range ops {
			prs = append(prs, &magistrala.AddPolicyReq{
				SubjectType:     "role",
				SubjectRelation: "member",
				Subject:         id,
				Relation:        op,
				Object:          entityID,
				ObjectType:      r.entityType,
			})
		}

		for _, member := range members {
			prs = append(prs, &magistrala.AddPolicyReq{
				SubjectType: "user",
				Subject:     auth.EncodeDomainUserID(domainID, member),
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

	if len(prs) > 0 {
		resp, err := r.auth.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})

		if err != nil {
			return []roles.RoleProvision{}, errors.Wrap(svcerr.ErrCreateEntity, err)
		}
		if !resp.Added {
			return []roles.RoleProvision{}, errors.Wrap(svcerr.ErrCreateEntity, errAddPolicies)
		}
		defer func() {
			if fnErr != nil {
				if errRollBack := r.addPolicyRollback(ctx, prs); errRollBack != nil {
					fnErr = errors.Wrap(fnErr, errors.Wrap(errRollbackRoles, errRollBack))
				}
			}
		}()
	}

	if _, err := r.repo.AddRoles(ctx, newRolesProvision); err != nil {
		return []roles.RoleProvision{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	return newRolesProvision, nil
}

func (r RolesSvc) addPolicyRollback(ctx context.Context, prs []*magistrala.AddPolicyReq) error {
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

func (r RolesSvc) RemoveNewEntityRoles(ctx context.Context, userID, domainID, entityID string, optionalPolicies []roles.OptionalPolicy, newRolesProvision []roles.RoleProvision) error {
	prs := []*magistrala.DeletePolicyReq{}

	roleIDs := []string{}
	for _, rp := range newRolesProvision {
		for _, op := range rp.OptionalOperations {
			prs = append(prs, &magistrala.DeletePolicyReq{
				SubjectType:     "role",
				SubjectRelation: "member",
				Subject:         rp.ID,
				Relation:        op,
				Object:          entityID,
				ObjectType:      r.entityType,
			})
		}

		for _, member := range rp.OptionalMembers {
			prs = append(prs, &magistrala.DeletePolicyReq{
				SubjectType: "user",
				Subject:     auth.EncodeDomainUserID(domainID, member),
				Relation:    "member",
				Object:      rp.ID,
				ObjectType:  "role",
			})
		}
		roleIDs = append(roleIDs, rp.ID)
	}

	if len(optionalPolicies) > 0 {
		for _, policy := range optionalPolicies {
			prs = append(prs, &magistrala.DeletePolicyReq{
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

	if len(prs) > 0 {
		resp, err := r.auth.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: prs})

		if err != nil {
			return errors.Wrap(svcerr.ErrDeletePolicies, err)
		}
		if !resp.Deleted {
			return svcerr.ErrDeletePolicies
		}

	}

	if err := r.repo.RemoveRoles(ctx, roleIDs); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	return nil
}

func (r RolesSvc) identify(ctx context.Context, token string) (identity, error) {
	resp, err := r.auth.Identify(ctx, &magistrala.IdentityReq{Token: token})
	if err != nil {
		return identity{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	return identity{ID: resp.GetId(), DomainID: resp.GetDomainId(), UserID: resp.GetUserId()}, nil
}
func (r RolesSvc) authorize(ctx context.Context, pr *magistrala.AuthorizeReq) error {
	resp, err := r.auth.Authorize(ctx, pr)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !resp.Authorized {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}
