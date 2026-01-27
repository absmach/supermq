// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"context"
)

type PolicyReq struct {
	// Domain contains the domain ID.
	Domain string `json:"domain,omitempty"`

	// Subject contains the subject ID or Token.
	Subject string `json:"subject"`

	// SubjectType contains the subject type. Supported subject types are
	// platform, group, domain, client, users.
	SubjectType string `json:"subject_type"`

	// SubjectKind contains the subject kind. Supported subject kinds are
	// token, users, platform, clients,  channels, groups, domain.
	SubjectKind string `json:"subject_kind"`

	// SubjectRelation contains subject relations.
	SubjectRelation string `json:"subject_relation,omitempty"`

	// Object contains the object ID.
	Object string `json:"object"`

	// ObjectKind contains the object kind. Supported object kinds are
	// users, platform, clients,  channels, groups, domain.
	ObjectKind string `json:"object_kind"`

	// ObjectType contains the object type. Supported object types are
	// platform, group, domain, client, users.
	ObjectType string `json:"object_type"`

	// Relation contains the relation. Supported relations are administrator, editor, contributor, member, guest, parent_group,group,domain.
	Relation string `json:"relation,omitempty"`

	// Permission contains the permission. Supported permissions are admin, delete, edit, share, view,
	// membership, create, admin_only, edit_only, view_only, membership_only, ext_admin, ext_edit, ext_view.
	Permission string `json:"permission,omitempty"`

	// PAT authorization fields
	UserID     string `json:"user_id,omitempty"`
	PatID      string `json:"pat_id,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
	DomainID   string `json:"domain_id,omitempty"`
	Operation  string `json:"operation,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
}

// Authz is supermq authorization library.
type Authorization interface {
	Authorize(ctx context.Context, pr PolicyReq) error
}
