// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package standalone

import (
	"context"

	"github.com/mainflux/mainflux/clients/policies"
	"github.com/mainflux/mainflux/pkg/errors"
	"google.golang.org/grpc"
)

var errUnsupported = errors.New("not supported in standalone mode")

var _ policies.AuthServiceClient = (*singleUserRepo)(nil)

type singleUserRepo struct {
	email string
	token string
}

// NewAuthService creates single user repository for constrained environments.
func NewAuthService(email, token string) policies.AuthServiceClient {
	return singleUserRepo{
		email: email,
		token: token,
	}
}

func (repo singleUserRepo) Issue(ctx context.Context, req *policies.IssueReq, opts ...grpc.CallOption) (*policies.Token, error) {
	if repo.token != req.GetEmail() {
		return nil, errors.ErrAuthentication
	}

	return &policies.Token{Value: repo.token}, nil
}

func (repo singleUserRepo) Identify(ctx context.Context, token *policies.Token, opts ...grpc.CallOption) (*policies.UserIdentity, error) {
	if repo.token != token.GetValue() {
		return nil, errors.ErrAuthentication
	}

	return &policies.UserIdentity{Id: repo.email, Email: repo.email}, nil
}

func (repo singleUserRepo) Authorize(ctx context.Context, req *policies.AuthorizeReq, _ ...grpc.CallOption) (r *policies.AuthorizeRes, err error) {
	if repo.email != req.Sub {
		return &policies.AuthorizeRes{}, errUnsupported
	}
	return &policies.AuthorizeRes{Authorized: true}, nil
}

func (repo singleUserRepo) AddPolicy(ctx context.Context, req *policies.AddPolicyReq, opts ...grpc.CallOption) (*policies.AddPolicyRes, error) {
	if repo.email != req.Sub {
		return &policies.AddPolicyRes{}, errUnsupported
	}
	return &policies.AddPolicyRes{Authorized: true}, nil
}

func (repo singleUserRepo) DeletePolicy(ctx context.Context, req *policies.DeletePolicyReq, opts ...grpc.CallOption) (*policies.DeletePolicyRes, error) {
	if repo.email != req.Sub {
		return &policies.DeletePolicyRes{}, errUnsupported
	}
	return &policies.DeletePolicyRes{Deleted: true}, nil
}

func (repo singleUserRepo) ListPolicies(ctx context.Context, in *policies.ListPoliciesReq, opts ...grpc.CallOption) (*policies.ListPoliciesRes, error) {
	return &policies.ListPoliciesRes{}, errUnsupported
}
