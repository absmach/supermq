// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	acl "github.com/ory/keto/proto/ory/keto/acl/v1alpha1"
)

// Authz represents a authorization service. It exposes
// functionalities through `auth` to perform authorization.
type Authz interface {
	// Authorize checks authorization of the given `subject`. Basically,
	// Authorize verifies that Is `subject` allowed to `relation` on
	// `object`. Authorize returns the bool indicating the authorization
	// of the given subject and the error. For example, the response as:
	// `false, nil` means that `subject` has not `relation` on the `object`.
	// Therefore, this response means that incoming request is denied.
	Authorize(ctx context.Context, subject, object, relation string) error

	// AddPolicy creates a policy for the given subject. So that, after
	// AddPolicy, `subject`has a `relation` on `object`. Returns non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, subject, object, relation string) error
}

// PolicyAgent facilitates the communication to authorization
// services and implements Authz functionalities for certain
// authorization services (e.g. ORY Keto).
type PolicyAgent interface {
	// CheckPolicy checks if the subject has a relation on the object.
	// It returns AuthorizationResult consisting of one boolean which indicates
	// result of the authorization (true: Allowed, false: Denied) and error.
	CheckPolicy(ctx context.Context, subject, object, relation string) (AuthorizationResult, error)

	// AddPolicy creates a policy for the given subject. So that, after
	// AddPolicy, `subject` has a `relation` on `object`. Returns non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, subject, object, relation string) error
}

// AuthorizationResult shows the result of the authorization check operations.
// AuthzError field is nil when the operation is allowed. If AuthzError is
// non-nil, the operation is denied.
type AuthorizationResult struct {
	AuthzError error
}

type policyAgent struct {
	writer  acl.WriteServiceClient
	checker acl.CheckServiceClient
}

// NewPolicyAgent returns a gRPC communication functionalities
// to communicate with the authorization services, e.g. ORY Keto.
func NewPolicyAgent(checker acl.CheckServiceClient, writer acl.WriteServiceClient) PolicyAgent {
	return policyAgent{checker: checker, writer: writer}
}

func (c policyAgent) CheckPolicy(ctx context.Context, subject, object, relation string) (AuthorizationResult, error) {
	res, err := c.checker.Check(context.Background(), &acl.CheckRequest{
		Namespace: ketoNamespace,
		Object:    object,
		Relation:  relation,
		Subject: &acl.Subject{Ref: &acl.Subject_Id{
			Id: subject,
		}},
	})
	if err != nil {
		return AuthorizationResult{err}, err
	}
	if !res.GetAllowed() {
		return AuthorizationResult{ErrAuthorization}, err
	}

	return AuthorizationResult{}, nil
}

func (c policyAgent) AddPolicy(ctx context.Context, subject, object, relation string) error {
	trt := c.writer.TransactRelationTuples
	_, err := trt(context.Background(), &acl.TransactRelationTuplesRequest{
		RelationTupleDeltas: []*acl.RelationTupleDelta{
			{
				Action: acl.RelationTupleDelta_INSERT,
				RelationTuple: &acl.RelationTuple{
					Namespace: ketoNamespace,
					Object:    object,
					Relation:  relation,
					Subject: &acl.Subject{Ref: &acl.Subject_Id{
						Id: subject,
					}},
				},
			},
		},
	})
	return err
}
