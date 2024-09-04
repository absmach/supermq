package domains

import (
	"context"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/pkg/domains"
	"github.com/absmach/magistrala/pkg/entityroles"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
)

var (
	errCreateDomainPolicy = errors.New("failed to create domain policy")
	errRollbackPolicy     = errors.New("failed to rollback policy")
	errRemovePolicyEngine = errors.New("failed to remove from policy engine")
)

type identity struct {
	ID       string
	DomainID string
	UserID   string
}
type service struct {
	repo       domains.DomainsRepository
	auth       magistrala.AuthServiceClient
	idProvider magistrala.IDProvider
	entityroles.RolesSvc
}

var _ domains.Service = (*service)(nil)

func New(repo domains.DomainsRepository, authClient magistrala.AuthServiceClient, idProvider magistrala.IDProvider, sidProvider magistrala.IDProvider) domains.Service {
	rolesSvc := entityroles.NewRolesSvc(auth.DomainType, repo, sidProvider, authClient, domains.AllowedOperations(), domains.BuiltInRoles())
	return &service{
		repo:       repo,
		auth:       authClient,
		idProvider: idProvider,
		RolesSvc:   rolesSvc,
	}
}

func (svc service) CreateDomain(ctx context.Context, token string, d domains.Domain) (do domains.Domain, err error) {
	user, err := svc.identify(ctx, token)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	d.CreatedBy = user.UserID

	domainID, err := svc.idProvider.ID()
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	d.ID = domainID

	if d.Status != domains.DisabledStatus && d.Status != domains.EnabledStatus {
		return domains.Domain{}, svcerr.ErrInvalidStatus
	}

	d.CreatedAt = time.Now()

	if err := svc.createDomainPolicy(ctx, user.UserID, domainID, auth.AdministratorRelation); err != nil {
		return domains.Domain{}, errors.Wrap(errCreateDomainPolicy, err)
	}
	defer func() {
		if err != nil {
			if errRollBack := svc.createDomainPolicyRollback(ctx, user.UserID, domainID, auth.AdministratorRelation); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackPolicy, errRollBack))
			}
		}
	}()
	dom, err := svc.repo.Save(ctx, d)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	return dom, nil
}

func (svc service) RetrieveDomain(ctx context.Context, token, id string) (domains.Domain, error) {
	user, err := svc.identify(ctx, token)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	domain, err := svc.repo.RetrieveByID(ctx, id)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	resp, err := svc.auth.Authorize(ctx, &magistrala.AuthorizeReq{
		Subject:     user.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      id,
		ObjectType:  auth.DomainType,
		Permission:  domains.ReadPermission,
	})
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !resp.Authorized {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthorization, err)

	}
	return domain, nil
}

func (svc service) UpdateDomain(ctx context.Context, token, id string, d domains.DomainReq) (domains.Domain, error) {
	user, err := svc.identify(ctx, token)
	if err != nil {
		return domains.Domain{}, err
	}
	resp, err := svc.auth.Authorize(ctx, &magistrala.AuthorizeReq{
		Subject:     user.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      id,
		ObjectType:  auth.DomainType,
		Permission:  domains.UpdatePermission,
	})
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !resp.Authorized {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthorization, err)

	}

	dom, err := svc.repo.Update(ctx, id, user.UserID, d)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return dom, nil
}

func (svc service) ChangeDomainStatus(ctx context.Context, token, id string, d domains.DomainReq) (domains.Domain, error) {
	user, err := svc.identify(ctx, token)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	resp, err := svc.auth.Authorize(ctx, &magistrala.AuthorizeReq{
		Subject:     user.ID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Object:      id,
		ObjectType:  auth.DomainType,
		Permission:  domains.UpdatePermission,
	})
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !resp.Authorized {
		return domains.Domain{}, errors.Wrap(svcerr.ErrAuthorization, err)

	}

	dom, err := svc.repo.Update(ctx, id, user.UserID, d)
	if err != nil {
		return domains.Domain{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return dom, nil
}

func (svc service) ListDomains(ctx context.Context, token string, p domains.Page) (domains.DomainsPage, error) {
	user, err := svc.identify(ctx, token)
	if err != nil {
		return domains.DomainsPage{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	p.SubjectID = user.ID
	//ToDo : Check list without below function and confirm and decide to remove or not
	// if _, err := svc.auth.Authorize(ctx, &magistrala.AuthorizeReq{
	// 	Subject:     user.ID,
	// 	SubjectType: auth.UserType,
	// 	Permission:  auth.AdminPermission,
	// 	ObjectType:  auth.PlatformType,
	// 	Object:      auth.MagistralaObject,
	// }); err == nil {
	// 	p.SubjectID = ""
	// }

	dp, err := svc.repo.ListDomains(ctx, p)
	if err != nil {
		return domains.DomainsPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return dp, nil
}

func (svc service) createDomainPolicy(ctx context.Context, userID, domainID, relation string) (err error) {
	// prs := []PolicyReq{
	// 	{
	// 		Subject:     EncodeDomainUserID(domainID, userID),
	// 		SubjectType: UserType,
	// 		SubjectKind: UsersKind,
	// 		Relation:    relation,
	// 		Object:      domainID,
	// 		ObjectType:  DomainType,
	// 	},
	// 	{
	// 		Subject:     MagistralaObject,
	// 		SubjectType: PlatformType,
	// 		Relation:    PlatformRelation,
	// 		Object:      domainID,
	// 		ObjectType:  DomainType,
	// 	},
	// }

	return nil
}

func (svc service) createDomainPolicyRollback(ctx context.Context, userID, domainID, relation string) error {
	// prs := []PolicyReq{
	// 	{
	// 		Subject:     EncodeDomainUserID(domainID, userID),
	// 		SubjectType: UserType,
	// 		SubjectKind: UsersKind,
	// 		Relation:    relation,
	// 		Object:      domainID,
	// 		ObjectType:  DomainType,
	// 	},
	// 	{
	// 		Subject:     MagistralaObject,
	// 		SubjectType: PlatformType,
	// 		Relation:    PlatformRelation,
	// 		Object:      domainID,
	// 		ObjectType:  DomainType,
	// 	},
	// }
	// if errPolicy := svc.agent.DeletePolicies(ctx, prs); errPolicy != nil {
	// 	err = errors.Wrap(errRemovePolicyEngine, errPolicy)
	// }

	return nil
}

func (svc service) identify(ctx context.Context, token string) (identity, error) {
	resp, err := svc.auth.Identify(ctx, &magistrala.IdentityReq{Token: token})
	if err != nil {
		return identity{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	return identity{ID: resp.GetId(), DomainID: resp.GetDomainId(), UserID: resp.GetUserId()}, nil
}
