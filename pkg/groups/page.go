// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package groups

import "github.com/absmach/magistrala/things"

// PageMeta contains page metadata that helps navigation.
type PageMeta struct {
	Total    uint64        `json:"total"`
	Offset   uint64        `json:"offset"`
	Limit    uint64        `json:"limit"`
	Name     string        `json:"name,omitempty"`
	ID       string        `json:"id,omitempty"`
	DomainID string        `json:"domain_id,omitempty"`
	Tag      string        `json:"tag,omitempty"`
	Metadata Metadata      `json:"metadata,omitempty"`
	Status   things.Status `json:"status,omitempty"`
}
