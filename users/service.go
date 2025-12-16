// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/absmach/supermq"
	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	apiutil "github.com/absmach/supermq/api/http/util"
	smqauth "github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/gofrs/uuid/v5"
)

const defaultUsernamePrefix = "user"

var (
	errIssueToken                   = errors.NewServiceError("failed to issue token")
	errRecoveryToken                = errors.NewServiceError("failed to generate password recovery token")
	errLoginDisableUser             = errors.NewServiceError("failed to login in disabled user")
	errAuthorizedCreateUser         = errors.NewAuthNError("not authorized to create user")
	errRetrieveUserVerification     = errors.NewServiceError("failed to retrieve user verification")
	errCreateUserVerification       = errors.NewServiceError("failed to create user verification")
	errAddUserVerification          = errors.NewServiceError("failed to add user verification")
	errEncodeUserVerification       = errors.NewServiceError("failed to encode user verification")
	errSendUserVerification         = errors.NewServiceError("failed to send user verification email")
	errMatchUserVerification        = errors.NewServiceError("user verification does not match with stored verification")
	errValidateUserVerification     = errors.NewServiceError("failed to validate user verification")
	errUpdateUserVerification       = errors.NewServiceError("failed to update user verification")
	errUpdateUserVerificationStatus = errors.NewServiceError("failed to update user verification status")
	errRetrieveUserFromDB           = errors.NewServiceError("failed to retrieve user with provided id from database")
	errAuthorizedListUsers          = errors.NewAuthNError("not authorized to list users")
	errListUsersFromDB              = errors.NewServiceError("failed to list users from database")
	errSearchUsersFromDB            = errors.NewServiceError("failed to search users from database")
	errAuthorizedUpdateUser         = errors.NewAuthNError("not authorized to update user details")
	errUpdateUserDB                 = errors.NewServiceError("failed to update user details in database")
	errSimilarUpdateEmail           = errors.NewConflictError("new email is similar to the current email")
	errRetrieveUserByEmail          = errors.NewServiceError("failed to retrieve user with provided email from database")
	errSendPasswordReset            = errors.NewServiceError("failed to send password reset email")
	errHashingSecret                = errors.NewServiceError("failed to hash user secret")
	errAuthorizedUpdateUserRole     = errors.NewAuthNError("not authorized to update user role")
	errUpdateUserRole               = errors.NewServiceError("failed to update user role")
	errDeleteUser                   = errors.NewServiceError("failed to delete user")

	usernameRegExp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{34}[a-z0-9]$`)
)

type service struct {
	token      grpcTokenV1.TokenServiceClient
	users      Repository
	idProvider supermq.IDProvider
	policies   policies.Service
	hasher     Hasher
	email      Emailer
}

// NewService returns a new Users service implementation.
func NewService(token grpcTokenV1.TokenServiceClient, urepo Repository, policyService policies.Service, emailer Emailer, hasher Hasher, idp supermq.IDProvider) Service {
	return service{
		token:      token,
		users:      urepo,
		policies:   policyService,
		hasher:     hasher,
		email:      emailer,
		idProvider: idp,
	}
}

func (svc service) Register(ctx context.Context, session authn.Session, u User, selfRegister bool) (uc User, err error) {
	if !selfRegister {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{}, errAuthorizedCreateUser.Embed(err)
		}
	}

	userID, err := svc.idProvider.ID()
	if err != nil {
		return User{}, svcerr.ErrIssueProviderID.Embed(err)
	}

	if u.Credentials.Secret != "" {
		hash, err := svc.hasher.Hash(u.Credentials.Secret)
		if err != nil {
			return User{}, svcerr.ErrHashPassword.Embed(err)
		}
		u.Credentials.Secret = hash
	}

	if u.Status != DisabledStatus && u.Status != EnabledStatus {
		return User{}, svcerr.ErrInvalidStatus
	}
	if u.Role != UserRole && u.Role != AdminRole {
		return User{}, svcerr.ErrInvalidRole
	}
	u.ID = userID
	u.CreatedAt = time.Now().UTC()

	if err := svc.addUserPolicy(ctx, u.ID, u.Role); err != nil {
		return User{}, svcerr.ErrAddPolicies.Embed(err)
	}
	defer func() {
		if err != nil {
			if errRollback := svc.addUserPolicyRollback(ctx, u.ID, u.Role); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()
	user, err := svc.users.Save(ctx, u)
	if err != nil {
		return User{}, svcerr.ErrFailedToSaveEntityDB.Embed(err)
	}
	return user, nil
}

func (svc service) SendVerification(ctx context.Context, session authn.Session) error {
	dbUser, err := svc.users.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return errRetrieveUserFromDB.Embed(err)
	}

	if !dbUser.VerifiedAt.IsZero() {
		return svcerr.ErrUserAlreadyVerified.Embed(err)
	}

	uv, err := svc.users.RetrieveUserVerification(ctx, dbUser.ID, dbUser.Email)
	if err != nil && err != repoerr.ErrNotFound {
		return errRetrieveUserVerification.Embed(err)
	}

	if err = uv.Valid(); err != nil {
		uv, err = NewUserVerification(dbUser.ID, dbUser.Email)
		if err != nil {
			return errCreateUserVerification.Embed(err)
		}
		if err := svc.users.AddUserVerification(ctx, uv); err != nil {
			return errAddUserVerification.Embed(err)
		}
	}

	uvs, err := uv.Encode()
	if err != nil {
		return errEncodeUserVerification.Embed(err)
	}

	if err := svc.email.SendVerification([]string{dbUser.Email}, dbUser.Credentials.Username, uvs); err != nil {
		return errSendUserVerification.Embed(err)
	}
	return nil
}

func (svc service) VerifyEmail(ctx context.Context, token string) (User, error) {
	var received UserVerification
	if err := received.Decode(token); err != nil {
		return User{}, svcerr.ErrInvalidUserVerification.Embed(err)
	}

	stored, err := svc.users.RetrieveUserVerification(ctx, received.UserID, received.Email)
	if err != nil {
		return User{}, errRetrieveUserVerification.Embed(err)
	}

	if err := stored.Match(received); err != nil {
		return User{}, errMatchUserVerification.Embed(err)
	}

	if err := stored.Valid(); err != nil {
		if err == svcerr.ErrUserVerificationExpired {
			return User{}, svcerr.ErrUserVerificationExpired.Embed(err)
		}
		return User{}, errValidateUserVerification.Embed(err)
	}

	stored.UsedAt = time.Now().UTC()
	if err = svc.users.UpdateUserVerification(ctx, stored); err != nil {
		return User{}, errUpdateUserVerification.Embed(err)
	}

	user := User{
		ID:         stored.UserID,
		Email:      stored.Email,
		VerifiedAt: time.Now().UTC(),
	}
	user, err = svc.users.UpdateVerifiedAt(ctx, user)
	if err == repoerr.ErrNotFound {
		return User{}, svcerr.ErrInvalidUserVerification.Embed(err)
	}
	if err != nil {
		return User{}, errUpdateUserVerificationStatus.Embed(err)
	}

	return user, nil
}

func (svc service) IssueToken(ctx context.Context, identity, secret string) (*grpcTokenV1.Token, error) {
	var dbUser User
	var err error

	if _, parseErr := mail.ParseAddress(identity); parseErr != nil {
		dbUser, err = svc.users.RetrieveByUsername(ctx, identity)
	} else {
		dbUser, err = svc.users.RetrieveByEmail(ctx, identity)
	}

	if err == repoerr.ErrNotFound {
		return &grpcTokenV1.Token{}, svcerr.ErrLogin.Embed(err)
	}

	if err != nil {
		return &grpcTokenV1.Token{}, svcerr.ErrAuthentication.Embed(err)
	}

	if err := svc.hasher.Compare(secret, dbUser.Credentials.Secret); err != nil {
		return &grpcTokenV1.Token{}, svcerr.ErrLogin.Embed(err)
	}

	token, err := svc.token.Issue(ctx, &grpcTokenV1.IssueReq{UserId: dbUser.ID, UserRole: uint32(dbUser.Role + 1), Type: uint32(smqauth.AccessKey), Verified: !dbUser.VerifiedAt.IsZero()})
	if err != nil {
		return &grpcTokenV1.Token{}, errIssueToken.Embed(err)
	}

	return token, nil
}

func (svc service) RefreshToken(ctx context.Context, session authn.Session, refreshToken string) (*grpcTokenV1.Token, error) {
	dbUser, err := svc.users.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return &grpcTokenV1.Token{}, errRetrieveUserFromDB.Embed(err)
	}
	if dbUser.Status == DisabledStatus {
		return &grpcTokenV1.Token{}, errLoginDisableUser.Embed(err)
	}

	token, err := svc.token.Refresh(ctx, &grpcTokenV1.RefreshReq{RefreshToken: refreshToken, Verified: !dbUser.VerifiedAt.IsZero()})
	if err != nil {
		return &grpcTokenV1.Token{}, errIssueToken.Embed(err)
	}

	return token, nil
}

func (svc service) View(ctx context.Context, session authn.Session, id string) (User, error) {
	user, err := svc.users.RetrieveByID(ctx, id)
	if err != nil {
		return User{}, errRetrieveUserFromDB.Embed(err)
	}

	if session.UserID != id {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				ID:          user.ID,
				Credentials: Credentials{Username: user.Credentials.Username},
			}, nil
		}
	}

	user.Credentials.Secret = ""

	return user, nil
}

func (svc service) ViewProfile(ctx context.Context, session authn.Session) (User, error) {
	user, err := svc.users.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return User{}, errRetrieveUserFromDB.Embed(err)
	}
	user.Credentials.Secret = ""

	return user, nil
}

func (svc service) ListUsers(ctx context.Context, session authn.Session, pm Page) (UsersPage, error) {
	if err := svc.checkSuperAdmin(ctx, session); err != nil {
		return UsersPage{}, errAuthorizedListUsers.Embed(err)
	}

	pm.Role = AllRole
	pg, err := svc.users.RetrieveAll(ctx, pm)
	if err != nil {
		return UsersPage{}, errListUsersFromDB.Embed(err)
	}
	return pg, nil
}

func (svc service) SearchUsers(ctx context.Context, pm Page) (UsersPage, error) {
	page := Page{
		Offset:    pm.Offset,
		Limit:     pm.Limit,
		FirstName: pm.FirstName,
		LastName:  pm.LastName,
		Username:  pm.Username,
		Id:        pm.Id,
		Role:      UserRole,
	}

	cp, err := svc.users.SearchUsers(ctx, page)
	if err != nil {
		return UsersPage{}, errSearchUsersFromDB.Embed(err)
	}

	return cp, nil
}

func (svc service) Update(ctx context.Context, session authn.Session, id string, usr UserReq) (User, error) {
	if session.UserID != id {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{}, errAuthorizedUpdateUser.Embed(err)
		}
	}
	u, err := svc.users.RetrieveByID(ctx, id)
	if err != nil {
		return User{}, errRetrieveUserFromDB.Embed(err)
	}
	if u.AuthProvider != "" {
		if changed(usr.FirstName, u.FirstName) ||
			changed(usr.LastName, u.LastName) ||
			changed(usr.ProfilePicture, u.ProfilePicture) {
			return User{}, svcerr.ErrExternalAuthProviderCouldNotUpdate.Embed(err)
		}
	}
	updatedAt := time.Now().UTC()
	usr.UpdatedAt = &updatedAt
	usr.UpdatedBy = &session.UserID

	user, err := svc.users.Update(ctx, id, usr)
	if err != nil {
		return User{}, errUpdateUserDB.Embed(err)
	}
	return user, nil
}

func (svc service) UpdateTags(ctx context.Context, session authn.Session, id string, usr UserReq) (User, error) {
	if session.UserID != id {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{}, errAuthorizedUpdateUser.Embed(err)
		}
	}

	updatedAt := time.Now().UTC()
	usr.UpdatedAt = &updatedAt
	usr.UpdatedBy = &session.UserID

	user, err := svc.users.Update(ctx, id, usr)
	if err != nil {
		return User{}, errUpdateUserDB.Embed(err)
	}

	return user, nil
}

func (svc service) UpdateProfilePicture(ctx context.Context, session authn.Session, id string, usr UserReq) (User, error) {
	if session.UserID != id {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{}, errAuthorizedUpdateUser.Embed(err)
		}
	}

	u, err := svc.users.RetrieveByID(ctx, id)
	if err != nil {
		return User{}, errRetrieveUserFromDB.Embed(err)
	}
	if u.AuthProvider != "" {
		return User{}, svcerr.ErrExternalAuthProviderCouldNotUpdate.Embed(err)
	}

	updatedAt := time.Now().UTC()
	usr.UpdatedAt = &updatedAt
	usr.UpdatedBy = &session.UserID

	user, err := svc.users.Update(ctx, id, usr)
	if err != nil {
		return User{}, errUpdateUserDB.Embed(err)
	}

	return user, nil
}

func (svc service) UpdateEmail(ctx context.Context, session authn.Session, userID, email string) (User, error) {
	if session.UserID != userID {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{}, errAuthorizedUpdateUser.Embed(err)
		}
	}
	oldUsr, err := svc.users.RetrieveByID(ctx, userID)
	if err != nil {
		return User{}, errRetrieveUserFromDB.Embed(err)
	}
	if oldUsr.AuthProvider != "" {
		return User{}, svcerr.ErrExternalAuthProviderCouldNotUpdate.Embed(err)
	}
	if oldUsr.Email == email {
		return User{}, errSimilarUpdateEmail.Embed(err)
	}

	usr := User{
		ID:         userID,
		Email:      email,
		UpdatedAt:  time.Now().UTC(),
		UpdatedBy:  session.UserID,
		VerifiedAt: time.Time{},
	}

	user, err := svc.users.UpdateEmail(ctx, usr)
	if err != nil {
		return User{}, errUpdateUserDB.Embed(err)
	}
	return user, nil
}

func (svc service) SendPasswordReset(ctx context.Context, email string) error {
	user, err := svc.users.RetrieveByEmail(ctx, email)
	if err != nil {
		return errRetrieveUserByEmail.Embed(err)
	}

	issueReq := &grpcTokenV1.IssueReq{
		UserId:   user.ID,
		UserRole: uint32(user.Role + 1),
		Type:     uint32(smqauth.RecoveryKey),
	}
	token, err := svc.token.Issue(ctx, issueReq)
	if err != nil {
		return errRecoveryToken.Embed(err)
	}

	if err := svc.email.SendPasswordReset([]string{email}, user.Credentials.Username, token.AccessToken); err != nil {
		return errSendPasswordReset.Embed(err)
	}

	return nil
}

func (svc service) ResetSecret(ctx context.Context, session authn.Session, secret string) error {
	u, err := svc.users.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return errRetrieveUserFromDB.Embed(err)
	}

	secret, err = svc.hasher.Hash(secret)
	if err != nil {
		return errHashingSecret.Embed(err)
	}
	u = User{
		ID:    u.ID,
		Email: u.Email,
		Credentials: Credentials{
			Secret: secret,
		},
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: session.UserID,
	}
	if _, err := svc.users.UpdateSecret(ctx, u); err != nil {
		return errUpdateUserDB.Embed(err)
	}
	return nil
}

func (svc service) UpdateSecret(ctx context.Context, session authn.Session, oldSecret, newSecret string) (User, error) {
	dbUser, err := svc.users.RetrieveByID(ctx, session.UserID)
	if err != nil {
		return User{}, errRetrieveUserFromDB.Embed(err)
	}
	if _, err := svc.IssueToken(ctx, dbUser.Credentials.Username, oldSecret); err != nil {
		return User{}, err
	}
	newSecret, err = svc.hasher.Hash(newSecret)
	if err != nil {
		return User{}, errHashingSecret.Embed(err)
	}
	dbUser.Credentials.Secret = newSecret
	dbUser.UpdatedAt = time.Now().UTC()
	dbUser.UpdatedBy = session.UserID

	dbUser, err = svc.users.UpdateSecret(ctx, dbUser)
	if err != nil {
		return User{}, errUpdateUserDB.Embed(err)
	}

	return dbUser, nil
}

func (svc service) UpdateUsername(ctx context.Context, session authn.Session, id, username string) (User, error) {
	if session.UserID != id {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{}, errAuthorizedUpdateUser.Embed(err)
		}
	}

	usr := User{
		ID: id,
		Credentials: Credentials{
			Username: username,
		},
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: session.UserID,
	}
	updatedUser, err := svc.users.UpdateUsername(ctx, usr)
	if err != nil {
		return User{}, errUpdateUserDB.Embed(err)
	}
	return updatedUser, nil
}

func (svc service) UpdateRole(ctx context.Context, session authn.Session, usr User) (User, error) {
	if err := svc.checkSuperAdmin(ctx, session); err != nil {
		return User{}, errAuthorizedUpdateUserRole.Embed(err)
	}
	usr = User{
		ID:        usr.ID,
		Role:      usr.Role,
		UpdatedAt: time.Now().UTC(),
		UpdatedBy: session.UserID,
	}

	if err := svc.updateUserPolicy(ctx, usr.ID, usr.Role); err != nil {
		return User{}, errUpdateUserRole.Embed(err)
	}

	u, err := svc.users.UpdateRole(ctx, usr)
	if err != nil {
		// If failed to update role in DB, then revert back to platform admin policies in spicedb
		if errRollback := svc.updateUserPolicy(ctx, usr.ID, UserRole); errRollback != nil {
			return User{}, errors.NewServiceError(errRollback.Error())
		}
		return User{}, errUpdateUserRole.Embed(err)
	}
	return u, nil
}

func (svc service) Enable(ctx context.Context, session authn.Session, id string) (User, error) {
	u := User{
		ID:        id,
		UpdatedAt: time.Now().UTC(),
		Status:    EnabledStatus,
	}
	user, err := svc.changeUserStatus(ctx, session, u)
	if err != nil {
		return User{}, svcerr.ErrEnableUser.Embed(err)
	}

	return user, nil
}

func (svc service) Disable(ctx context.Context, session authn.Session, id string) (User, error) {
	user := User{
		ID:        id,
		UpdatedAt: time.Now().UTC(),
		Status:    DisabledStatus,
	}
	user, err := svc.changeUserStatus(ctx, session, user)
	if err != nil {
		return User{}, svcerr.ErrDisableUser.Embed(err)
	}

	return user, nil
}

func (svc service) changeUserStatus(ctx context.Context, session authn.Session, user User) (User, error) {
	if session.UserID != user.ID {
		if err := svc.checkSuperAdmin(ctx, session); err != nil {
			return User{}, err
		}
	}
	dbu, err := svc.users.RetrieveByID(ctx, user.ID)
	if err != nil {
		return User{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if dbu.Status == user.Status {
		return User{}, errors.ErrStatusAlreadyAssigned
	}
	user.UpdatedBy = session.UserID

	user, err = svc.users.ChangeStatus(ctx, user)
	if err != nil {
		return User{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return user, nil
}

func (svc service) Delete(ctx context.Context, session authn.Session, id string) error {
	user := User{
		ID:        id,
		UpdatedAt: time.Now().UTC(),
		Status:    DeletedStatus,
	}

	if _, err := svc.changeUserStatus(ctx, session, user); err != nil {
		return errDeleteUser.Embed(err)
	}

	return nil
}

func (svc *service) checkSuperAdmin(ctx context.Context, session authn.Session) error {
	if !session.SuperAdmin {
		if err := svc.users.CheckSuperAdmin(ctx, session.UserID); err != nil {
			return errors.Wrap(svcerr.ErrAuthorization, err)
		}
	}

	return nil
}

func (svc service) OAuthCallback(ctx context.Context, user User) (User, error) {
	u, err := svc.users.RetrieveByEmail(ctx, user.Email)

	if errors.Contains(err, repoerr.ErrNotFound) {
		user.Credentials.Username = generateUsername(user.Email)
		u, err = svc.Register(ctx, authn.Session{}, user, true)
		if err != nil {
			if errors.Contains(err, errors.ErrUsernameNotAvailable) {
				return User{}, errors.ErrTryAgain
			}
			return User{}, err
		}
	}

	if err != nil && !errors.Contains(err, repoerr.ErrNotFound) {
		return User{}, err
	}

	if u.VerifiedAt.IsZero() {
		user.ID = u.ID
		user.VerifiedAt = time.Now()
		u, err = svc.users.UpdateVerifiedAt(ctx, user)
		if err != nil {
			return User{}, err
		}
	}

	return User{ID: u.ID, Role: u.Role, VerifiedAt: u.VerifiedAt}, nil
}

func (svc service) OAuthAddUserPolicy(ctx context.Context, user User) error {
	return svc.addUserPolicy(ctx, user.ID, user.Role)
}

func (svc service) Identify(ctx context.Context, session authn.Session) (string, error) {
	return session.UserID, nil
}

func (svc service) addUserPolicy(ctx context.Context, userID string, role Role) error {
	policyList := []policies.Policy{}

	policyList = append(policyList, policies.Policy{
		SubjectType: policies.UserType,
		Subject:     userID,
		Relation:    policies.MemberRelation,
		ObjectType:  policies.PlatformType,
		Object:      policies.SuperMQObject,
	})

	if role == AdminRole {
		policyList = append(policyList, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.SuperMQObject,
		})
	}
	err := svc.policies.AddPolicies(ctx, policyList)
	if err != nil {
		return errors.Wrap(svcerr.ErrAddPolicies, err)
	}

	return nil
}

func (svc service) addUserPolicyRollback(ctx context.Context, userID string, role Role) error {
	policyList := []policies.Policy{}

	policyList = append(policyList, policies.Policy{
		SubjectType: policies.UserType,
		Subject:     userID,
		Relation:    policies.MemberRelation,
		ObjectType:  policies.PlatformType,
		Object:      policies.SuperMQObject,
	})

	if role == AdminRole {
		policyList = append(policyList, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.SuperMQObject,
		})
	}
	err := svc.policies.DeletePolicies(ctx, policyList)
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}

	return nil
}

func (svc service) updateUserPolicy(ctx context.Context, userID string, role Role) error {
	switch role {
	case AdminRole:
		err := svc.policies.AddPolicy(ctx, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.SuperMQObject,
		})
		if err != nil {
			return errors.Wrap(svcerr.ErrAddPolicies, err)
		}

		return nil
	case UserRole:
		fallthrough
	default:
		err := svc.policies.DeletePolicyFilter(ctx, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			ObjectType:  policies.PlatformType,
			Object:      policies.SuperMQObject,
		})
		if err != nil {
			return errors.Wrap(svcerr.ErrDeletePolicies, err)
		}

		return nil
	}
}

func generateUsername(email string) string {
	uniqueSuffix := generateRandomID()
	emailPrefix := extractEmailPrefix(email)
	return fmt.Sprintf("%s_%s", emailPrefix, uniqueSuffix)
}

func extractEmailPrefix(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return defaultUsernamePrefix
	}

	prefix := parts[0]
	cleaned := usernameRegExp.ReplaceAllString(prefix, "")

	cleaned = sanitizeForUsername(cleaned, 15)
	if cleaned == "" {
		cleaned = defaultUsernamePrefix
	}

	return cleaned
}

func generateRandomID() string {
	// Generate 8 random bytes (will result in 16 hex chars, truncated to 10)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback: use UUID if crypto/rand fails (should never happen)
		id, uuidErr := uuid.NewV4()
		if uuidErr != nil {
			// Last resort fallback
			return fmt.Sprintf("%x", time.Now().UnixNano())[:10]
		}
		return hex.EncodeToString(id.Bytes())[:10]
	}
	return hex.EncodeToString(randomBytes)[:10]
}

// sanitizeForUsername extracts and cleans a string for use in username generation.
// As per the username requirements:
// - It keeps only lowercase alphanumeric characters, hyphens, and underscores
// - ensures valid boundaries (no hyphens/underscores at start/end)
// - removes consecutive hyphens/underscores (to pass validation)
// and finally limits the result to maxLen characters.
func sanitizeForUsername(s string, maxLen int) string {
	if s == "" {
		return ""
	}

	// Convert to lowercase
	s = strings.ToLower(s)

	// Filter characters - keep only alphanumeric, hyphen, underscore
	buf := make([]byte, 0, len(s))
	var lastChar byte

	for i := 0; i < len(s); i++ {
		c := s[i]

		// Keep alphanumeric, hyphen, underscore
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			// Skip if current char is hyphen/underscore and same as last char
			if (c == '-' || c == '_') && c == lastChar {
				continue // Skip consecutive hyphens or underscores
			}
			buf = append(buf, c)
			lastChar = c
		} else {
			lastChar = 0 // Reset on special char
		}
	}

	cleaned := string(buf)

	// Trim invalid boundary characters
	cleaned = strings.Trim(cleaned, "-_")

	// Limit length
	if len(cleaned) > maxLen {
		cleaned = cleaned[:maxLen]
		// Re-trim in case truncation exposed hyphen/underscore at end
		cleaned = strings.TrimRight(cleaned, "-_")
	}

	return cleaned
}

func changed(updated *string, old string) bool {
	if updated == nil {
		return false
	}

	return *updated != old
}
