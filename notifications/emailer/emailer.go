// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"context"
	"fmt"

	grpcUsersV1 "github.com/absmach/supermq/api/grpc/users/v1"
	"github.com/absmach/supermq/internal/email"
	"github.com/absmach/supermq/notifications"
	"github.com/absmach/supermq/pkg/errors"
)

var (
	errFetchingUser = errors.New("failed to fetch user information")
	errSendingEmail = errors.New("failed to send email")
)

var _ notifications.Notifier = (*notifier)(nil)

type notifier struct {
	usersClient     grpcUsersV1.UsersServiceClient
	invitationAgent *email.Agent
	acceptanceAgent *email.Agent
	rejectionAgent  *email.Agent
	fromAddress     string
	fromName        string
}

// Config represents the emailer configuration.
type Config struct {
	FromAddress         string
	FromName            string
	InvitationTemplate  string
	AcceptanceTemplate  string
	RejectionTemplate   string
	EmailHost           string
	EmailPort           string
	EmailUsername       string
	EmailPassword       string
}

// New creates a new email notifier.
func New(usersClient grpcUsersV1.UsersServiceClient, cfg Config) (notifications.Notifier, error) {
	invitationEmailCfg := &email.Config{
		Host:        cfg.EmailHost,
		Port:        cfg.EmailPort,
		Username:    cfg.EmailUsername,
		Password:    cfg.EmailPassword,
		FromAddress: cfg.FromAddress,
		FromName:    cfg.FromName,
		Template:    cfg.InvitationTemplate,
	}
	invitationAgent, err := email.New(invitationEmailCfg)
	if err != nil {
		return nil, err
	}

	acceptanceEmailCfg := &email.Config{
		Host:        cfg.EmailHost,
		Port:        cfg.EmailPort,
		Username:    cfg.EmailUsername,
		Password:    cfg.EmailPassword,
		FromAddress: cfg.FromAddress,
		FromName:    cfg.FromName,
		Template:    cfg.AcceptanceTemplate,
	}
	acceptanceAgent, err := email.New(acceptanceEmailCfg)
	if err != nil {
		return nil, err
	}

	rejectionEmailCfg := &email.Config{
		Host:        cfg.EmailHost,
		Port:        cfg.EmailPort,
		Username:    cfg.EmailUsername,
		Password:    cfg.EmailPassword,
		FromAddress: cfg.FromAddress,
		FromName:    cfg.FromName,
		Template:    cfg.RejectionTemplate,
	}
	rejectionAgent, err := email.New(rejectionEmailCfg)
	if err != nil {
		return nil, err
	}

	return &notifier{
		usersClient:     usersClient,
		invitationAgent: invitationAgent,
		acceptanceAgent: acceptanceAgent,
		rejectionAgent:  rejectionAgent,
		fromAddress:     cfg.FromAddress,
		fromName:        cfg.FromName,
	}, nil
}

func (n *notifier) SendInvitationNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	users, err := n.fetchUsers(ctx, []string{inviterID, inviteeID})
	if err != nil {
		return errors.Wrap(errFetchingUser, err)
	}

	inviter, ok := users[inviterID]
	if !ok {
		return errors.Wrap(errFetchingUser, fmt.Errorf("inviter not found: %s", inviterID))
	}

	invitee, ok := users[inviteeID]
	if !ok {
		return errors.Wrap(errFetchingUser, fmt.Errorf("invitee not found: %s", inviteeID))
	}

	inviterName := n.getUserDisplayName(inviter)
	inviteeName := n.getUserDisplayName(invitee)

	if domainName == "" {
		domainName = domainID
	}
	if roleName == "" {
		roleName = roleID
	}

	subject := "Domain Invitation"
	content := fmt.Sprintf("%s has invited you to join the domain '%s' as '%s'.", inviterName, domainName, roleName)

	if err := n.invitationAgent.Send([]string{invitee.Email}, "", subject, "", inviteeName, content, n.fromName); err != nil {
		return errors.Wrap(errSendingEmail, err)
	}

	return nil
}

func (n *notifier) SendAcceptanceNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	users, err := n.fetchUsers(ctx, []string{inviterID, inviteeID})
	if err != nil {
		return errors.Wrap(errFetchingUser, err)
	}

	inviter, ok := users[inviterID]
	if !ok {
		return errors.Wrap(errFetchingUser, fmt.Errorf("inviter not found: %s", inviterID))
	}

	invitee, ok := users[inviteeID]
	if !ok {
		return errors.Wrap(errFetchingUser, fmt.Errorf("invitee not found: %s", inviteeID))
	}

	inviterName := n.getUserDisplayName(inviter)
	inviteeName := n.getUserDisplayName(invitee)

	if domainName == "" {
		domainName = domainID
	}
	if roleName == "" {
		roleName = roleID
	}

	subject := "Invitation Accepted"
	content := fmt.Sprintf("%s has accepted your invitation to join the domain '%s' as '%s'.", inviteeName, domainName, roleName)

	if err := n.acceptanceAgent.Send([]string{inviter.Email}, "", subject, "", inviterName, content, n.fromName); err != nil {
		return errors.Wrap(errSendingEmail, err)
	}

	return nil
}

func (n *notifier) SendRejectionNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	users, err := n.fetchUsers(ctx, []string{inviterID, inviteeID})
	if err != nil {
		return errors.Wrap(errFetchingUser, err)
	}

	inviter, ok := users[inviterID]
	if !ok {
		return errors.Wrap(errFetchingUser, fmt.Errorf("inviter not found: %s", inviterID))
	}

	invitee, ok := users[inviteeID]
	if !ok {
		return errors.Wrap(errFetchingUser, fmt.Errorf("invitee not found: %s", inviteeID))
	}

	inviterName := n.getUserDisplayName(inviter)
	inviteeName := n.getUserDisplayName(invitee)

	if domainName == "" {
		domainName = domainID
	}
	if roleName == "" {
		roleName = roleID
	}

	subject := "Invitation Declined"
	content := fmt.Sprintf("%s has declined your invitation to join the domain '%s' as '%s'.", inviteeName, domainName, roleName)

	if err := n.rejectionAgent.Send([]string{inviter.Email}, "", subject, "", inviterName, content, n.fromName); err != nil {
		return errors.Wrap(errSendingEmail, err)
	}

	return nil
}

func (n *notifier) fetchUsers(ctx context.Context, userIDs []string) (map[string]*grpcUsersV1.User, error) {
	req := &grpcUsersV1.RetrieveUsersReq{
		Ids:    userIDs,
		Limit:  uint64(len(userIDs)),
		Offset: 0,
	}

	res, err := n.usersClient.RetrieveUsers(ctx, req)
	if err != nil {
		return nil, err
	}

	users := make(map[string]*grpcUsersV1.User)
	for _, user := range res.Users {
		users[user.Id] = user
	}

	return users, nil
}

func (n *notifier) getUserDisplayName(user *grpcUsersV1.User) string {
	if user.FirstName != "" && user.LastName != "" {
		return fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	}
	if user.FirstName != "" {
		return user.FirstName
	}
	if user.Username != "" {
		return user.Username
	}
	if user.Email != "" {
		return user.Email
	}
	return user.Id
}
