// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auditlogs

import (
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/pkg/errors"
)

type EntityType uint8

const (
	UserEntity EntityType = iota
	GroupEntity
	ClientEntity
	ChannelEntity
	MessageEntity
)

// String representation of the possible entity type values.
const (
	userEntityType    = "user"
	groupEntityType   = "group"
	clientEntityType  = "client"
	channelEntityType = "channel"
	messageEntityType = "message"
)

// String converts entity type to string literal.
func (e EntityType) String() string {
	switch e {
	case UserEntity:
		return userEntityType
	case GroupEntity:
		return groupEntityType
	case ClientEntity:
		return clientEntityType
	case ChannelEntity:
		return channelEntityType
	case MessageEntity:
		return messageEntityType
	default:
		return ""
	}
}

// ToEntityType converts string value to a valid entity type.
func ToEntityType(entityType string) (EntityType, error) {
	switch entityType {
	case userEntityType:
		return UserEntity, nil
	case groupEntityType:
		return GroupEntity, nil
	case clientEntityType:
		return ClientEntity, nil
	case channelEntityType:
		return ChannelEntity, nil
	case messageEntityType:
		return MessageEntity, nil
	default:
		return EntityType(0), apiutil.ErrInvalidEntityType
	}
}

type EntityState uint8

const (
	CreatedState EntityState = iota
	UpdatedState
	EnabledState
	DisabledState
	DeletedState
)

const (
	createdState  = "created"
	updatedState  = "updated"
	enabledState  = "enabled"
	disabledState = "disabled"
	deletedState  = "deleted"
)

func (s EntityState) String() string {
	switch s {
	case CreatedState:
		return createdState
	case UpdatedState:
		return updatedState
	case EnabledState:
		return enabledState
	case DisabledState:
		return disabledState
	case DeletedState:
		return deletedState
	default:
		return ""
	}
}

func ToEntityState(state string) (EntityState, error) {
	switch state {
	case createdState:
		return CreatedState, nil
	case updatedState:
		return UpdatedState, nil
	case deletedState:
		return DeletedState, nil
	default:
		return EntityState(0), errors.ErrMalformedEntity
	}
}
