// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"time"

	"github.com/absmach/supermq/channels"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/events"
)

const (
	channelPrefix       = "channels."
	channelCreate       = channelPrefix + "create"
	channelUpdate       = channelPrefix + "update"
	channelChangeStatus = channelPrefix + "change_status"
	channelRemove       = channelPrefix + "remove"
	channelView         = channelPrefix + "view"
	channelList         = channelPrefix + "list"
	channelConnect      = channelPrefix + "connect"
	channelDisconnect   = channelPrefix + "disconnect"
	channelSetParent    = channelPrefix + "set_parent"
	channelRemoveParent = channelPrefix + "remove_parent"
)

var (
	_ events.Event = (*createChannelEvent)(nil)
	_ events.Event = (*updateChannelEvent)(nil)
	_ events.Event = (*changeStatusChannelEvent)(nil)
	_ events.Event = (*viewChannelEvent)(nil)
	_ events.Event = (*listChannelEvent)(nil)
	_ events.Event = (*removeChannelEvent)(nil)
	_ events.Event = (*connectEvent)(nil)
	_ events.Event = (*disconnectEvent)(nil)
)

type createChannelEvent struct {
	channels.Channel
	domainID string
}

func (cce createChannelEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation":  channelCreate,
		"id":         cce.ID,
		"status":     cce.Status.String(),
		"created_at": cce.CreatedAt,
		"domain":     cce.domainID,
	}

	if cce.Name != "" {
		val["name"] = cce.Name
	}
	if len(cce.Tags) > 0 {
		val["tags"] = cce.Tags
	}
	if cce.Metadata != nil {
		val["metadata"] = cce.Metadata
	}

	return val, nil
}

type updateChannelEvent struct {
	channels.Channel
	operation string
	domainID  string
}

func (uce updateChannelEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation":  channelUpdate,
		"updated_at": uce.UpdatedAt,
		"updated_by": uce.UpdatedBy,
		"domain":     uce.domainID,
	}
	if uce.operation != "" {
		val["operation"] = channelUpdate + "_" + uce.operation
	}

	if uce.ID != "" {
		val["id"] = uce.ID
	}
	if uce.Name != "" {
		val["name"] = uce.Name
	}
	if len(uce.Tags) > 0 {
		val["tags"] = uce.Tags
	}
	if uce.Metadata != nil {
		val["metadata"] = uce.Metadata
	}
	if !uce.CreatedAt.IsZero() {
		val["created_at"] = uce.CreatedAt
	}
	if uce.Status.String() != "" {
		val["status"] = uce.Status.String()
	}

	return val, nil
}

type changeStatusChannelEvent struct {
	id        string
	status    string
	updatedAt time.Time
	updatedBy string
	domainID  string
}

func (rce changeStatusChannelEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"operation":  channelChangeStatus,
		"id":         rce.id,
		"status":     rce.status,
		"updated_at": rce.updatedAt,
		"updated_by": rce.updatedBy,
		"domain":     rce.domainID,
	}, nil
}

type viewChannelEvent struct {
	channels.Channel
	domainID string
}

func (vce viewChannelEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": channelView,
		"id":        vce.ID,
		"domain":    vce.domainID,
	}

	if vce.Name != "" {
		val["name"] = vce.Name
	}
	if len(vce.Tags) > 0 {
		val["tags"] = vce.Tags
	}
	if vce.Metadata != nil {
		val["metadata"] = vce.Metadata
	}
	if !vce.CreatedAt.IsZero() {
		val["created_at"] = vce.CreatedAt
	}
	if !vce.UpdatedAt.IsZero() {
		val["updated_at"] = vce.UpdatedAt
	}
	if vce.UpdatedBy != "" {
		val["updated_by"] = vce.UpdatedBy
	}
	if vce.Status.String() != "" {
		val["status"] = vce.Status.String()
	}

	return val, nil
}

type listChannelEvent struct {
	channels.PageMetadata
	domainID string
}

func (lce listChannelEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": channelList,
		"total":     lce.Total,
		"offset":    lce.Offset,
		"limit":     lce.Limit,
		"domain":    lce.domainID,
	}

	if lce.Name != "" {
		val["name"] = lce.Name
	}
	if lce.Order != "" {
		val["order"] = lce.Order
	}
	if lce.Dir != "" {
		val["dir"] = lce.Dir
	}
	if lce.Metadata != nil {
		val["metadata"] = lce.Metadata
	}
	if lce.Tag != "" {
		val["tag"] = lce.Tag
	}
	if lce.Permission != "" {
		val["permission"] = lce.Permission
	}
	if lce.Status.String() != "" {
		val["status"] = lce.Status.String()
	}
	if len(lce.IDs) > 0 {
		val["ids"] = lce.IDs
	}

	return val, nil
}

type listChannelByClientEvent struct {
	domainID string
	clientID string
	channels.PageMetadata
}

func (lcte listChannelByClientEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": channelList,
		"client_id": lcte.clientID,
		"total":     lcte.Total,
		"offset":    lcte.Offset,
		"limit":     lcte.Limit,
		"domain":    lcte.domainID,
	}

	if lcte.Name != "" {
		val["name"] = lcte.Name
	}
	if lcte.Order != "" {
		val["order"] = lcte.Order
	}
	if lcte.Dir != "" {
		val["dir"] = lcte.Dir
	}
	if lcte.Metadata != nil {
		val["metadata"] = lcte.Metadata
	}
	if lcte.Tag != "" {
		val["tag"] = lcte.Tag
	}
	if lcte.Permission != "" {
		val["permission"] = lcte.Permission
	}
	if lcte.Status.String() != "" {
		val["status"] = lcte.Status.String()
	}
	if len(lcte.IDs) > 0 {
		val["ids"] = lcte.IDs
	}

	return val, nil
}

type removeChannelEvent struct {
	id       string
	domainID string
}

func (dce removeChannelEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"operation": channelRemove,
		"id":        dce.id,
		"domain":    dce.domainID,
	}, nil
}

type connectEvent struct {
	domainID string
	chIDs    []string
	thIDs    []string
	types    []connections.ConnType
}

func (ce connectEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"operation":   channelConnect,
		"client_ids":  ce.thIDs,
		"channel_ids": ce.chIDs,
		"types":       ce.types,
		"domain":      ce.domainID,
	}, nil
}

type disconnectEvent struct {
	domainID string
	chIDs    []string
	thIDs    []string
	types    []connections.ConnType
}

func (de disconnectEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"operation":   channelDisconnect,
		"client_ids":  de.thIDs,
		"channel_ids": de.chIDs,
		"types":       de.types,
		"domain":      de.domainID,
	}, nil
}

type setParentGroupEvent struct {
	id            string
	parentGroupID string
	domainID      string
}

func (spge setParentGroupEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"operation":       channelSetParent,
		"id":              spge.id,
		"parent_group_id": spge.parentGroupID,
		"domain":          spge.domainID,
	}, nil
}

type removeParentGroupEvent struct {
	id       string
	domainID string
}

func (rpge removeParentGroupEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"operation": channelRemoveParent,
		"id":        rpge.id,
		"domain":    rpge.domainID,
	}, nil
}
