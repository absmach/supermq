// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"context"
)

// Invitation represents minimal invitation information needed for notifications.
type Invitation struct {
	InvitedBy     string `db:"invited_by"`
	InviteeUserID string `db:"invitee_user_id"`
	DomainID      string `db:"domain_id"`
}

// InvitationRepository provides access to invitation information.
type InvitationRepository interface {
	// RetrieveInvitation retrieves an invitation by invitee user ID and domain ID.
	RetrieveInvitation(ctx context.Context, inviteeUserID, domainID string) (Invitation, error)
}
