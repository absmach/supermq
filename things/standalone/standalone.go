// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package standalone

import (
	"context"

	"github.com/absmach/magistrala"
	grpcclient "github.com/absmach/magistrala/auth/api/grpc"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/policy"
	"google.golang.org/grpc"
)

var (
	_ grpcclient.AuthServiceClient = (*singleUserAuth)(nil)
	_ policy.PolicyService         = (*singleUserPolicyClient)(nil)
)

type singleUserAuth struct {
	id    string
	token string
}

// NewAuthService creates single user repository for constrained environments.
func NewAuthService(id, token string) grpcclient.AuthServiceClient {
	return singleUserAuth{
		id:    id,
		token: token,
	}
}

func (repo singleUserAuth) Login(ctx context.Context, in *magistrala.IssueReq, opts ...grpc.CallOption) (*magistrala.Token, error) {
	return nil, nil
}

func (repo singleUserAuth) Refresh(ctx context.Context, in *magistrala.RefreshReq, opts ...grpc.CallOption) (*magistrala.Token, error) {
	return nil, nil
}

func (repo singleUserAuth) Issue(ctx context.Context, in *magistrala.IssueReq, opts ...grpc.CallOption) (*magistrala.Token, error) {
	return nil, nil
}

func (repo singleUserAuth) Identify(ctx context.Context, in *magistrala.IdentityReq, opts ...grpc.CallOption) (*magistrala.IdentityRes, error) {
	if repo.token != in.GetToken() {
		return nil, svcerr.ErrAuthentication
	}

	return &magistrala.IdentityRes{Id: repo.id}, nil
}

func (repo singleUserAuth) Authorize(ctx context.Context, in *magistrala.AuthorizeReq, opts ...grpc.CallOption) (*magistrala.AuthorizeRes, error) {
	if repo.id != in.Subject {
		return &magistrala.AuthorizeRes{Authorized: false}, svcerr.ErrAuthorization
	}

	return &magistrala.AuthorizeRes{Authorized: true}, nil
}

type singleUserPolicyClient struct {
	id    string
	token string
}

// NewPolicyService creates single user policy service for constrained environments.
func NewPolicyService(id, token string) policy.PolicyService {
	return singleUserPolicyClient{
		id:    id,
		token: token,
	}
}

func (repo singleUserPolicyClient) AddPolicy(ctx context.Context, pr policy.PolicyReq) error {
	return nil
}

func (repo singleUserPolicyClient) AddPolicies(ctx context.Context, prs []policy.PolicyReq) error {
	return nil
}

func (repo singleUserPolicyClient) DeletePolicyFilter(ctx context.Context, pr policy.PolicyReq) error {
	return nil
}

func (repo singleUserPolicyClient) DeletePolicies(ctx context.Context, prs []policy.PolicyReq) error {
	return nil
}

func (repo singleUserPolicyClient) ListObjects(ctx context.Context, pr policy.PolicyReq, nextPageToken string, limit uint64) (policy.PolicyPage, error) {
	return policy.PolicyPage{}, nil
}

func (repo singleUserPolicyClient) ListAllObjects(ctx context.Context, pr policy.PolicyReq) (policy.PolicyPage, error) {
	return policy.PolicyPage{}, nil
}

func (repo singleUserPolicyClient) CountObjects(ctx context.Context, pr policy.PolicyReq) (uint64, error) {
	return 0, nil
}

func (repo singleUserPolicyClient) ListSubjects(ctx context.Context, pr policy.PolicyReq, nextPageToken string, limit uint64) (policy.PolicyPage, error) {
	return policy.PolicyPage{}, nil
}

func (repo singleUserPolicyClient) ListAllSubjects(ctx context.Context, pr policy.PolicyReq) (policy.PolicyPage, error) {
	return policy.PolicyPage{}, nil
}

func (repo singleUserPolicyClient) CountSubjects(ctx context.Context, pr policy.PolicyReq) (uint64, error) {
	return 0, nil
}

func (repo singleUserPolicyClient) ListPermissions(ctx context.Context, pr policy.PolicyReq, permissionsFilter []string) (policy.Permissions, error) {
	return nil, nil
}

func (repo singleUserPolicyClient) DeleteEntityPolicies(ctx context.Context, entityType, id string) error {
	return nil
}
