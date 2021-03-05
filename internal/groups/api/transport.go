package groups

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-zoo/bone"
	intapihttp "github.com/mainflux/mainflux/internal/api/http"
	"github.com/mainflux/mainflux/internal/groups"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	maxNameSize = 254
	offsetKey   = "offset"
	limitKey    = "limit"
	nameKey     = "name"
	levelKey    = "level"
	metadataKey = "metadata"
	treeKey     = "tree"
	contentType = "application/json"

	defOffset = 0
	defLimit  = 10
	defLevel  = 1
)

func DecodeListGroupsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, groups.ErrUnsupportedContentType
	}

	l, err := intapihttp.ReadUintQuery(r, levelKey, defLevel)
	if err != nil {
		return nil, err
	}

	n, err := intapihttp.ReadStringQuery(r, nameKey)
	if err != nil {
		return nil, err
	}

	m, err := intapihttp.ReadMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	t, err := intapihttp.ReadBoolQuery(r, treeKey)
	if err != nil {
		return nil, err
	}

	req := listGroupsReq{
		token:    r.Header.Get("Authorization"),
		level:    l,
		name:     n,
		metadata: m,
		tree:     t,
		groupID:  bone.GetValue(r, "groupID"),
	}
	return req, nil
}

func DecodeListMemberGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, groups.ErrUnsupportedContentType
	}

	o, err := intapihttp.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := intapihttp.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	n, err := intapihttp.ReadStringQuery(r, nameKey)
	if err != nil {
		return nil, err
	}

	m, err := intapihttp.ReadMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	t, err := intapihttp.ReadBoolQuery(r, treeKey)
	if err != nil {
		return nil, err
	}

	req := listMemberGroupReq{
		token:    r.Header.Get("Authorization"),
		groupID:  bone.GetValue(r, "groupID"),
		memberID: bone.GetValue(r, "memberID"),
		offset:   o,
		limit:    l,
		name:     n,
		metadata: m,
		tree:     t,
	}
	return req, nil
}

func DecodeGroupCreate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, groups.ErrUnsupportedContentType
	}

	var req createGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(groups.ErrFailedDecode, err)
	}

	req.token = r.Header.Get("Authorization")
	return req, nil
}

func DecodeGroupUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, groups.ErrUnsupportedContentType
	}

	var req updateGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(groups.ErrFailedDecode, err)
	}

	req.id = bone.GetValue(r, "groupID")
	req.token = r.Header.Get("Authorization")
	return req, nil
}

func DecodeGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := groupReq{
		token:   r.Header.Get("Authorization"),
		groupID: bone.GetValue(r, "groupID"),
		name:    bone.GetValue(r, "name"),
	}

	return req, nil
}

func DecodeMemberGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := memberGroupReq{
		token:    r.Header.Get("Authorization"),
		groupID:  bone.GetValue(r, "groupID"),
		memberID: bone.GetValue(r, "memberID"),
	}

	return req, nil
}
