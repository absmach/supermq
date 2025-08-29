// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	smqsdk "github.com/absmach/supermq/pkg/sdk"
	smqusers "github.com/absmach/supermq/users"
	"github.com/spf13/cobra"
)

const (
	token                = "token"
	refreshtoken         = "refreshtoken"
	profile              = "profile"
	resetpasswordrequest = "resetpasswordrequest"
	resetpassword        = "resetpassword"
	password             = "password"
	search               = "search"
	username             = "username"
	email                = "email"
	role                 = "role"

	// Usage strings for user operations.
	usageUserCreate           = "cli users create <first_name> <last_name> <email> <username> <password> [user_auth_token]"
	usageUserGet              = "cli users <user_id|all> get <user_auth_token>"
	usageUserToken            = "cli users token <username> <password>"
	usageUserRefreshToken     = "cli users refreshtoken <token>"
	usageUserUpdate           = "cli users <user_id> update <JSON_string> <user_auth_token>"
	usageUserUpdateTags       = "cli users <user_id> update tags <tags> <user_auth_token>"
	usageUserUpdateUsername   = "cli users <user_id> update username <username> <user_auth_token>"
	usageUserUpdateEmail      = "cli users <user_id> update email <email> <user_auth_token>"
	usageUserUpdateRole       = "cli users <user_id> update role <role> <user_auth_token>"
	usageUserProfile          = "cli users profile <user_auth_token>"
	usageUserResetPasswordReq = "cli users resetpasswordrequest <email>"
	usageUserResetPassword    = "cli users resetpassword <password> <confpass> <password_request_token>"
	usageUserPassword         = "cli users password <old_password> <password> <user_auth_token>"
	usageUserEnable           = "cli users <user_id> enable <user_auth_token>"
	usageUserDisable          = "cli users <user_id> disable <user_auth_token>"
	usageUserDelete           = "cli users <user_id> delete <user_auth_token>"
	usageUserSearch           = "cli users search <query> <user_auth_token>"
)

func NewUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users <user_id_or_all> <operation> [args...]",
		Short: "Users management",
		Long: `Format: <user_id|all> <operation> [additional_args...]

Examples:
  users all get <user_auth_token>                                       					# Get all entities
  users <user_id> get <user_auth_token>                                 					# Get specific entity
  users <user_id> update <JSON_string> <user_auth_token>                					# Update entity
  users <user_id> update tags <tags> <user_auth_token>                  					# Update entity tags
  users <user_id> update username <username> <user_auth_token>          					# Update username
  users <user_id> update email <email> <user_auth_token>                					# Update email
  users <user_id> enable <user_auth_token>                              					# Enable entity
  users <user_id> disable <user_auth_token>                             					# Disable entity
  users <user_id> delete <user_auth_token>                              					# Delete entity
  users create <first_name> <last_name> <email> <username> <password> [user_auth_token]  	# Create entity
  users token <username> <password>                                     					# Get token
  users profile <user_auth_token>                                       					# Get entity profile
  users search <query> <user_auth_token>                                					# Search entities`,

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			switch args[0] {
			case create:
				handleUserCreate(cmd, args[1:])
				return
			case token:
				handleUserToken(cmd, args[1], args[2:])
				return
			case refreshtoken:
				handleUserRefreshToken(cmd, args[1], args[2:])
				return
			case profile:
				handleUserProfile(cmd, args[1], args[2:])
				return
			case resetpasswordrequest:
				handleUserResetPasswordRequest(cmd, args[1], args[2:])
				return
			case resetpassword:
				handleUserResetPassword(cmd, args[1], args[2:])
				return
			case password:
				handleUserPassword(cmd, args[1], args[2:])
				return
			case search:
				handleUserSearch(cmd, args[1], args[2:])
				return
			}

			userParams := args[0]
			operation := args[1]
			opArgs := args[2:]

			switch operation {
			case get:
				handleUserGet(cmd, userParams, opArgs)
			case update:
				handleUserUpdate(cmd, userParams, opArgs)
			case enable:
				handleUserEnable(cmd, userParams, opArgs)
			case disable:
				handleUserDisable(cmd, userParams, opArgs)
			case delete:
				handleUserDelete(cmd, userParams, opArgs)
			default:
				logErrorCmd(*cmd, fmt.Errorf("unknown operation: %s", operation))
			}
		},
	}

	return cmd
}

func handleUserCreate(cmd *cobra.Command, args []string) {
	if len(args) < 5 || len(args) > 6 {
		logUsageCmd(*cmd, usageUserCreate)
		return
	}
	if len(args) == 5 {
		args = append(args, "")
	}

	user := smqsdk.User{
		FirstName: args[0],
		LastName:  args[1],
		Email:     args[2],
		Credentials: smqsdk.Credentials{
			Username: args[3],
			Secret:   args[4],
		},
		Status: smqusers.EnabledStatus.String(),
	}
	user, err := sdk.CreateUser(cmd.Context(), user, args[5])
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, user)
}

func handleUserGet(cmd *cobra.Command, userParams string, args []string) {
	if len(args) != 1 {
		logUsageCmd(*cmd, usageUserGet)
		return
	}

	if userParams == all {
		metadata, err := convertMetadata(Metadata)
		if err != nil {
			logErrorCmd(*cmd, err)
			return
		}

		pageMetadata := smqsdk.PageMetadata{
			Username: Username,
			Identity: Identity,
			Offset:   Offset,
			Limit:    Limit,
			Metadata: metadata,
			Status:   Status,
		}

		l, err := sdk.Users(cmd.Context(), pageMetadata, args[0])
		if err != nil {
			logErrorCmd(*cmd, err)
			return
		}
		logJSONCmd(*cmd, l)
		return
	}

	u, err := sdk.User(cmd.Context(), userParams, args[0])
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, u)
}

func handleUserUpdate(cmd *cobra.Command, userID string, args []string) {
	if len(args) < 2 || len(args) > 3 {
		if len(args) >= 1 {
			switch args[0] {
			case tags:
				logUsageCmd(*cmd, usageUserUpdateTags)
				return
			case username:
				logUsageCmd(*cmd, usageUserUpdateUsername)
				return
			case email:
				logUsageCmd(*cmd, usageUserUpdateEmail)
				return
			case role:
				logUsageCmd(*cmd, usageUserUpdateRole)
				return
			}
		}
		logUsageCmd(*cmd, usageUserUpdate)
		return
	}

	var user smqsdk.User
	if args[0] == "tags" {
		if len(args) != 3 {
			logUsageCmd(*cmd, usageUserUpdateTags)
			return
		}
		if err := json.Unmarshal([]byte(args[1]), &user.Tags); err != nil {
			logErrorCmd(*cmd, err)
			return
		}
		user.ID = userID
		user, err := sdk.UpdateUserTags(cmd.Context(), user, args[2])
		if err != nil {
			logErrorCmd(*cmd, err)
			return
		}
		logJSONCmd(*cmd, user)
		return
	}

	if args[0] == "email" {
		if len(args) != 3 {
			logUsageCmd(*cmd, usageUserUpdateEmail)
			return
		}
		user.ID = userID
		user.Email = args[1]
		user, err := sdk.UpdateUserEmail(cmd.Context(), user, args[2])
		if err != nil {
			logErrorCmd(*cmd, err)
			return
		}
		logJSONCmd(*cmd, user)
		return
	}

	if args[0] == "username" {
		if len(args) != 3 {
			logUsageCmd(*cmd, usageUserUpdateUsername)
			return
		}
		user.ID = userID
		user.Credentials.Username = args[1]
		user, err := sdk.UpdateUsername(cmd.Context(), user, args[2])
		if err != nil {
			logErrorCmd(*cmd, err)
			return
		}
		logJSONCmd(*cmd, user)
		return
	}

	if args[0] == "role" {
		if len(args) != 3 {
			logUsageCmd(*cmd, usageUserUpdateRole)
			return
		}
		user.ID = userID
		user.Role = args[1]
		user, err := sdk.UpdateUserRole(cmd.Context(), user, args[2])
		if err != nil {
			logErrorCmd(*cmd, err)
			return
		}
		logJSONCmd(*cmd, user)
		return
	}

	if len(args) != 2 {
		logUsageCmd(*cmd, usageUserUpdate)
		return
	}

	if err := json.Unmarshal([]byte(args[0]), &user); err != nil {
		logErrorCmd(*cmd, err)
		return
	}
	user.ID = userID
	user, err := sdk.UpdateUser(cmd.Context(), user, args[1])
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, user)
}

func handleUserEnable(cmd *cobra.Command, userID string, args []string) {
	if len(args) != 1 {
		logUsageCmd(*cmd, usageUserEnable)
		return
	}

	user, err := sdk.EnableUser(cmd.Context(), userID, args[0])
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, user)
}

func handleUserDisable(cmd *cobra.Command, userID string, args []string) {
	if len(args) != 1 {
		logUsageCmd(*cmd, usageUserDisable)
		return
	}

	user, err := sdk.DisableUser(cmd.Context(), userID, args[0])
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, user)
}

func handleUserDelete(cmd *cobra.Command, userID string, args []string) {
	if len(args) != 1 {
		logUsageCmd(*cmd, usageUserDelete)
		return
	}

	if err := sdk.DeleteUser(cmd.Context(), userID, args[0]); err != nil {
		logErrorCmd(*cmd, err)
		return
	}
	logOKCmd(*cmd)
}

func handleUserToken(cmd *cobra.Command, username string, args []string) {
	if len(args) != 1 {
		logUsageCmd(*cmd, usageUserToken)
		return
	}

	loginReq := smqsdk.Login{
		Username: username,
		Password: args[0],
	}

	token, err := sdk.CreateToken(cmd.Context(), loginReq)
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, token)
}

func handleUserRefreshToken(cmd *cobra.Command, refreshToken string, args []string) {
	if len(args) != 0 {
		logUsageCmd(*cmd, usageUserRefreshToken)
		return
	}

	token, err := sdk.RefreshToken(cmd.Context(), refreshToken)
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, token)
}

func handleUserProfile(cmd *cobra.Command, token string, args []string) {
	if len(args) != 0 {
		logUsageCmd(*cmd, usageUserProfile)
		return
	}

	user, err := sdk.UserProfile(cmd.Context(), token)
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, user)
}

func handleUserResetPasswordRequest(cmd *cobra.Command, email string, args []string) {
	if len(args) != 0 {
		logUsageCmd(*cmd, usageUserResetPasswordReq)
		return
	}

	if err := sdk.ResetPasswordRequest(cmd.Context(), email); err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logOKCmd(*cmd)
}

func handleUserResetPassword(cmd *cobra.Command, password string, args []string) {
	if len(args) != 2 {
		logUsageCmd(*cmd, usageUserResetPassword)
		return
	}

	if err := sdk.ResetPassword(cmd.Context(), password, args[0], args[1]); err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logOKCmd(*cmd)
}

func handleUserPassword(cmd *cobra.Command, oldPassword string, args []string) {
	if len(args) != 2 {
		logUsageCmd(*cmd, usageUserPassword)
		return
	}

	user, err := sdk.UpdatePassword(cmd.Context(), oldPassword, args[0], args[1])
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, user)
}

func handleUserSearch(cmd *cobra.Command, query string, args []string) {
	if len(args) != 1 {
		logUsageCmd(*cmd, usageUserSearch)
		return
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		logErrorCmd(*cmd, fmt.Errorf("failed to parse query: %s", err))
		return
	}

	pm := smqsdk.PageMetadata{
		Offset: Offset,
		Limit:  Limit,
		Name:   values.Get("name"),
		ID:     values.Get("id"),
	}

	if off, err := strconv.Atoi(values.Get("offset")); err == nil {
		pm.Offset = uint64(off)
	}

	if lim, err := strconv.Atoi(values.Get("limit")); err == nil {
		pm.Limit = uint64(lim)
	}

	users, err := sdk.SearchUsers(cmd.Context(), pm, args[0])
	if err != nil {
		logErrorCmd(*cmd, err)
		return
	}

	logJSONCmd(*cmd, users)
}
