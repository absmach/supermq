// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"
	acl "github.com/ory/keto/proto/ory/keto/acl/v1alpha1"
)

// PolicyReq represents an argument struct for making a policy related
// function calls.
type PolicyReq struct {
	Subject  string
	Object   string
	Relation string
}

// Authz represents a authorization service. It exposes
// functionalities through `auth` to perform authorization.
type Authz interface {
	// Authorize checks authorization of the given `subject`. Basically,
	// Authorize verifies that Is `subject` allowed to `relation` on
	// `object`. Authorize returns a non-nil error if the subject has
	// no relation on the object (which simply means the operation is
	// denied).
	Authorize(ctx context.Context, pr PolicyReq) error

	// AddPolicy creates a policy for the given subject, so that, after
	// AddPolicy, `subject` has a `relation` on `object`. Returns a non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, pr PolicyReq) error
}

// PolicyAgent facilitates the communication to authorization
// services and implements Authz functionalities for certain
// authorization services (e.g. ORY Keto).
type PolicyAgent interface {
	// CheckPolicy checks if the subject has a relation on the object.
	// It returns a non-nil error if the subject has no relation on
	// the object (which simply means the operation is denied).
	CheckPolicy(ctx context.Context, pr PolicyReq) error

	// AddPolicy creates a policy for the given subject, so that, after
	// AddPolicy, `subject` has a `relation` on `object`. Returns a non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, pr PolicyReq) error
}

type policyAgent struct {
	writer  acl.WriteServiceClient
	checker acl.CheckServiceClient
}

// NewPolicyAgent returns a gRPC communication functionalities
// to communicate with ORY Keto.
func NewPolicyAgent(checker acl.CheckServiceClient, writer acl.WriteServiceClient) PolicyAgent {
	return policyAgent{checker: checker, writer: writer}
}

func (c policyAgent) CheckPolicy(ctx context.Context, pr PolicyReq) error {
	res, err := c.checker.Check(context.Background(), &acl.CheckRequest{
		Namespace: ketoNamespace,
		Object:    pr.Object,
		Relation:  pr.Relation,
		Subject: &acl.Subject{Ref: &acl.Subject_Id{
			Id: pr.Subject,
		}},
	})
	if err != nil {
		return errors.Wrap(err, ErrAuthorization)
	}
	if !res.GetAllowed() {
		return ErrAuthorization
	}
	return nil
}

func (c policyAgent) AddPolicy(ctx context.Context, pr PolicyReq) error {
	trt := c.writer.TransactRelationTuples
	_, err := trt(context.Background(), &acl.TransactRelationTuplesRequest{
		RelationTupleDeltas: []*acl.RelationTupleDelta{
			{
				Action: acl.RelationTupleDelta_INSERT,
				RelationTuple: &acl.RelationTuple{
					Namespace: ketoNamespace,
					Object:    pr.Object,
					Relation:  pr.Relation,
					Subject: &acl.Subject{Ref: &acl.Subject_Id{
						Id: pr.Subject,
					}},
				},
			},
		},
	})
	return err
}
