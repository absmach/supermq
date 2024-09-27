package entityroles

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
	grpcclient "github.com/absmach/magistrala/auth/api/grpc"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/roles"
	"github.com/absmach/magistrala/pkg/svcutil"
)

type identity struct {
	ID       string
	DomainID string
	UserID   string
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
	auth         grpcclient.AuthServiceClient
	policy       magistrala.PolicyServiceClient
	actions      []roles.Action
	builtInRoles map[roles.BuiltInRoleName][]roles.Action
	opp          svcutil.OperationPerm
}

func NewRolesSvc(entityType string, repo roles.Repository, idProvider magistrala.IDProvider, auth grpcclient.AuthServiceClient, policyClient magistrala.PolicyServiceClient, actions []roles.Action, builtInRoles map[roles.BuiltInRoleName][]roles.Action, opPerm map[svcutil.Operation]svcutil.Permission) (RolesSvc, error) {
	opp := roles.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(opPerm); err != nil {
		return RolesSvc{}, err
	}
	if err := opp.Validate(); err != nil {
		return RolesSvc{}, err
	}
	rolesSvc := RolesSvc{
		entityType:   entityType,
		repo:         repo,
		idProvider:   idProvider,
		auth:         auth,
		policy:       policyClient,
		actions:      actions,
		builtInRoles: builtInRoles,
		opp:          opp,
	}
	if err := rolesSvc.validate(); err != nil {
		return RolesSvc{}, err
	}
	return rolesSvc, nil
}

func (r RolesSvc) validate() error {
	if err := r.opp.Validate(); err != nil {
		return err
	}
	return nil
}

func toRolesActions(actions []string) []roles.Action {
	roActions := []roles.Action{}
	for _, action := range actions {
		roActions = append(roActions, roles.Action(action))
	}
	return roActions
}

func roleActionsToString(roActions []roles.Action) []string {
	actions := []string{}
	for _, roAction := range roActions {
		actions = append(actions, roAction.String())
	}
	return actions
}

func roleMembersToString(roMems []roles.Member) []string {
	mems := []string{}
	for _, roMem := range roMems {
		mems = append(mems, roMem.String())
	}
	return mems
}

func (r RolesSvc) isActionAllowed(action roles.Action) bool {
	for _, cap := range r.actions {
		if cap == action {
			return true
		}
	}
	return false
}
func (r RolesSvc) validateActions(actions []roles.Action) error {
	for _, ac := range actions {
		action := roles.Action(ac)
		if !r.isActionAllowed(action) {
			return errors.Wrap(svcerr.ErrMalformedEntity, fmt.Errorf("invalid action %s ", action))
		}
	}
	return nil
}

func (r RolesSvc) AddRole(ctx context.Context, token, entityID string, roleName string, optionalActions []string, optionalMembers []string) (fnRole roles.Role, fnErr error) {

	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return roles.Role{}, err
	}

	if err := r.authorize(ctx, roles.OpAddRole, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
	}); err != nil {
		return roles.Role{}, err
	}

	// ToDo: Research: Discuss: There an option to have id as entityID_roleName where in roleName all space are removed with _ and starts with letter and supports only alphanumeric, space and hyphen
	id, err := r.idProvider.ID()
	if err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	if err := r.validateActions(toRolesActions(optionalActions)); err != nil {
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
			OptionalActions: optionalActions,
			OptionalMembers: optionalMembers,
		},
	}
	prs := []*magistrala.AddPolicyReq{}

	for _, cap := range optionalActions {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType:     "role",
			SubjectRelation: "member",
			Subject:         id,
			Relation:        cap,
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
		resp, err := r.policy.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})

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

	if err := r.authorize(ctx, roles.OpRemoveRole, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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
	resp, err := r.policy.DeletePolicyFilter(ctx, req)
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

	if err := r.authorize(ctx, roles.OpUpdateRoleName, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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
	if err := r.authorize(ctx, roles.OpRetrieveRole, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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

	if err := r.authorize(ctx, roles.OpRetrieveAllRoles, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
	}); err != nil {
		return roles.RolePage{}, err
	}

	ros, err := r.repo.RetrieveAllRoles(ctx, entityID, limit, offset)
	if err != nil {
		return roles.RolePage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return ros, nil
}

func (r RolesSvc) ListAvailableActions(ctx context.Context, token string) ([]string, error) {
	if _, err := r.identify(ctx, token); err != nil {
		return []string{}, err
	}
	acts := []string{}
	for _, a := range r.actions {
		acts = append(acts, string(a))
	}
	return acts, nil
}

func (r RolesSvc) RoleAddActions(ctx context.Context, token, entityID, roleName string, actions []string) (fnActs []string, fnErr error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return []string{}, err
	}

	if err := r.authorize(ctx, roles.OpRoleAddActions, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
	}); err != nil {
		return []string{}, err
	}

	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}

	if len(actions) == 0 {
		return []string{}, svcerr.ErrMalformedEntity
	}

	if err := r.validateActions(toRolesActions(actions)); err != nil {
		return []string{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
	}

	prs := []*magistrala.AddPolicyReq{}
	for _, cap := range actions {
		prs = append(prs, &magistrala.AddPolicyReq{
			SubjectType:     "role",
			SubjectRelation: "member",
			Subject:         ro.ID,
			Relation:        cap,
			Object:          entityID,
			ObjectType:      r.entityType,
		})
	}

	resp, err := r.policy.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})
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

	resActs, err := r.repo.RoleAddActions(ctx, ro, actions)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	return resActs, nil
}

func (r RolesSvc) RoleListActions(ctx context.Context, token, entityID, roleName string) ([]string, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return []string{}, err
	}

	if err := r.authorize(ctx, roles.OpRoleListActions, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
	}); err != nil {
		return []string{}, err
	}

	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	acts, err := r.repo.RoleListActions(ctx, ro.ID)
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return acts, nil

}

func (r RolesSvc) RoleCheckActionsExists(ctx context.Context, token, entityID, roleName string, actions []string) (bool, error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return false, err
	}
	if err := r.authorize(ctx, roles.OpRoleCheckActionsExists, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
	}); err != nil {
		return false, err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return false, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	result, err := r.repo.RoleCheckActionsExists(ctx, ro.ID, actions)
	if err != nil {
		return true, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return result, nil
}

func (r RolesSvc) RoleRemoveActions(ctx context.Context, token, entityID, roleName string, actions []string) (err error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := r.authorize(ctx, roles.OpRoleRemoveActions, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
	}); err != nil {
		return err
	}
	ro, err := r.repo.RetrieveRoleByEntityIDAndName(ctx, entityID, roleName)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	if len(actions) == 0 {
		return svcerr.ErrMalformedEntity
	}

	prs := []*magistrala.DeletePolicyReq{}
	for _, op := range actions {
		prs = append(prs, &magistrala.DeletePolicyReq{
			SubjectType:     "role",
			SubjectRelation: "member",
			Subject:         ro.ID,
			Relation:        op,
			Object:          entityID,
			ObjectType:      r.entityType,
		})
	}

	resp, err := r.policy.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: prs})
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	if !resp.Deleted {
		return svcerr.ErrDeletePolicies
	}

	ro.UpdatedAt = time.Now()
	ro.UpdatedBy = userInfo.UserID
	if err := r.repo.RoleRemoveActions(ctx, ro, actions); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r RolesSvc) RoleRemoveAllActions(ctx context.Context, token, entityID, roleName string) error {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := r.authorize(ctx, roles.OpRoleRemoveAllActions, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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

	resp, err := r.policy.DeletePolicyFilter(ctx, prs)
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	if !resp.Deleted {
		return svcerr.ErrDeletePolicies
	}

	ro.UpdatedAt = time.Now()
	ro.UpdatedBy = userInfo.UserID

	if err := r.repo.RoleRemoveAllActions(ctx, ro); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	return nil
}

func (r RolesSvc) RoleAddMembers(ctx context.Context, token, entityID, roleName string, members []string) (fnMems []string, fnErr error) {
	userInfo, err := r.identify(ctx, token)
	if err != nil {
		return []string{}, err
	}
	if err := r.authorize(ctx, roles.OpRoleAddMembers, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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

	resp, err := r.policy.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})
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
	if err := r.authorize(ctx, roles.OpRoleListMembers, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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
	if err := r.authorize(ctx, roles.OpRoleCheckMembersExists, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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
	if err := r.authorize(ctx, roles.OpRoleRemoveMembers, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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

	resp, err := r.policy.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: prs})
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
	if err := r.authorize(ctx, roles.OpRoleRemoveAllMembers, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		Subject:     userInfo.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      entityID,
		ObjectType:  r.entityType,
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

	resp, err := r.policy.DeletePolicyFilter(ctx, prs)
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
		actions, ok := r.builtInRoles[defaultRole]
		if !ok {
			return []roles.RoleProvision{}, fmt.Errorf("default role %s not found in in-built roles", defaultRole)
		}

		// There an option to have id as entityID_roleName where in roleName all space are removed with _ and starts with letter and supports only alphanumeric, space and hyphen
		id, err := r.idProvider.ID()
		if err != nil {
			return []roles.RoleProvision{}, errors.Wrap(svcerr.ErrCreateEntity, err)
		}

		if err := r.validateActions(actions); err != nil {
			return []roles.RoleProvision{}, errors.Wrap(svcerr.ErrMalformedEntity, err)
		}

		members := roleMembersToString(defaultRoleMembers)
		caps := roleActionsToString(actions)

		newRolesProvision = append(newRolesProvision, roles.RoleProvision{
			Role: roles.Role{
				ID:        id,
				Name:      defaultRole.String(),
				EntityID:  entityID,
				CreatedAt: time.Now(),
				CreatedBy: userID,
			},
			OptionalActions: caps,
			OptionalMembers: members,
		})

		for _, cap := range caps {
			prs = append(prs, &magistrala.AddPolicyReq{
				SubjectType:     "role",
				SubjectRelation: "member",
				Subject:         id,
				Relation:        cap,
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
		resp, err := r.policy.AddPolicies(ctx, &magistrala.AddPoliciesReq{AddPoliciesReq: prs})

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
	resp, err := r.policy.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: delPrs})
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
		for _, cap := range rp.OptionalActions {
			prs = append(prs, &magistrala.DeletePolicyReq{
				SubjectType:     "role",
				SubjectRelation: "member",
				Subject:         rp.ID,
				Relation:        cap,
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
		resp, err := r.policy.DeletePolicies(ctx, &magistrala.DeletePoliciesReq{DeletePoliciesReq: prs})

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
func (r RolesSvc) authorize(ctx context.Context, op svcutil.Operation, pr *magistrala.AuthorizeReq) error {
	perm, err := r.opp.GetPermission(op)
	if err != nil {
		return err
	}

	pr.Permission = perm.String()

	resp, err := r.auth.Authorize(ctx, pr)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !resp.Authorized {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}
