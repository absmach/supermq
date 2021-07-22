// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	acl "github.com/ory/keto/proto/ory/keto/acl/v1alpha1"
)

// PolicyService represents a authorization service. It exposes
// functionalities through `auth` to perform authorization.
type PolicyService interface {
	// Authorize checks authorization of the given `subject`. Basically,
	// Authorize verifies that Is `subject` allowed to `relation` on
	// `object`. Authorize returns the bool indicating the authorization
	// of the given subject and the error. For example, the response as:
	// `false, nil` means that `subject` has not `relation` on the `object`.
	// Therefore, this response means that incoming request is denied.
	Authorize(ctx context.Context, subject, object, relation string) (bool, error)

	// AddPolicy creates a policy for the given subject. So that, after
	// AddPolicy, `subject`has a `relation` on `object`. Returns non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, subject, object, relation string) error
}

// PolicyCommunicator facilitates the communication to authorization
// services and implements PolicyService functionalities for certain
// authorization services (e.g. ORY Keto).
type PolicyCommunicator interface {
	// CheckPolicy checks if the subject has a relation on the object.
	// It returns PolicyResult consisting of one boolean which indicates
	// result of the authorization (true: Allowed, false: Denied) and error.
	CheckPolicy(ctx context.Context, subject, object, relation string) (PolicyResult, error)

	// AddPolicy creates a policy for the given subject. So that, after
	// AddPolicy, `subject` has a `relation` on `object`. Returns non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, subject, object, relation string) error
}

// PolicyResult shows the result of the authorization check operations.
// Authorized field is true when the operation is allowed. If Authorized
// is false, the operation is denied.
type PolicyResult struct {
	Authorized bool
}

// KetoConfig represents config variables and functions in order to establish
// a connection to Keto through gRPC. The config variable (WritePort and ReadPort)
// are read from the ./docker/.env file. If you want to modify these config variables,
// you can change them through that .env file (under docker/ from the root directory).
type KetoConfig struct {
	WritePort string
	ReadPort  string
	Checker   acl.CheckServiceClient
	Writer    acl.WriteServiceClient
}

type communicator struct {
	kc KetoConfig
}

// NewPolicyCommunicator returns a gRPC communication functionalities
// to communicate with the authorization services, e.g. ORY Keto.
func NewPolicyCommunicator(kc KetoConfig) PolicyCommunicator {
	return communicator{kc}
}

func (c communicator) CheckPolicy(ctx context.Context, subject, object, relation string) (PolicyResult, error) {
	res, err := c.kc.Checker.Check(context.Background(), &acl.CheckRequest{
		Namespace: ketoNamespace,
		Object:    object,
		Relation:  relation,
		Subject: &acl.Subject{Ref: &acl.Subject_Id{
			Id: subject,
		}},
	})
	if err != nil {
		return PolicyResult{}, err
	}

	return PolicyResult{res.Allowed}, nil
}

func (c communicator) AddPolicy(ctx context.Context, subject, object, relation string) error {
	_, err := c.kc.Writer.TransactRelationTuples(context.Background(), &acl.TransactRelationTuplesRequest{
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
	if err != nil {
		return err
	}
	return nil
}
