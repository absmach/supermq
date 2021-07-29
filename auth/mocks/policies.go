// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux/auth"
)

type MockSubjectSet struct {
	Object   string
	Relation string
}

type ketoMock struct {
	authzDB map[string][]MockSubjectSet
}

// NewKetoMock returns a mock service for Keto.
// This mock is not implemented yet.
func NewKetoMock(db map[string][]MockSubjectSet) auth.PolicyCommunicator {
	return &ketoMock{db}
}

func (k *ketoMock) CheckPolicy(ctx context.Context, subject, object, relation string) (auth.PolicyResult, error) {
	ssList := k.authzDB[subject]
	for _, ss := range ssList {
		if ss.Object == object && ss.Relation == relation {
			return auth.PolicyResult{Authorized: true}, nil
		}
	}
	return auth.PolicyResult{Authorized: false}, nil
}

func (k *ketoMock) AddPolicy(ctx context.Context, subject, object, relation string) error {
	k.authzDB[subject] = append(k.authzDB[subject], MockSubjectSet{Object: object, Relation: relation})
	return nil
}
