// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"

	"github.com/absmach/supermq/notifications"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
)

var _ notifications.InvitationRepository = (*invitationRepository)(nil)

type invitationRepository struct {
	db postgres.Database
}

// NewInvitationRepository instantiates a PostgreSQL implementation of invitation repository.
func NewInvitationRepository(db postgres.Database) notifications.InvitationRepository {
	return &invitationRepository{
		db: db,
	}
}

func (repo *invitationRepository) RetrieveInvitation(ctx context.Context, inviteeUserID, domainID string) (notifications.Invitation, error) {
	query := `SELECT invited_by, invitee_user_id, domain_id
	          FROM invitations
	          WHERE invitee_user_id = $1 AND domain_id = $2`

	var invitation notifications.Invitation
	if err := repo.db.QueryRowxContext(ctx, query, inviteeUserID, domainID).StructScan(&invitation); err != nil {
		return notifications.Invitation{}, errors.Wrap(repoerr.ErrNotFound, err)
	}

	return invitation, nil
}
