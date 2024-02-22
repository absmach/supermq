// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
)

const channelParts = 2

func (sdk mgSDK) SendMessage(chanName, msg, key string) errors.SDKError {
	chanNameParts := strings.SplitN(chanName, ".", channelParts)
	chanID := chanNameParts[0]
	subtopicPart := ""
	if len(chanNameParts) == channelParts {
		subtopicPart = fmt.Sprintf("/%s", strings.ReplaceAll(chanNameParts[1], ".", "/"))
	}

	url := fmt.Sprintf("%s/channels/%s/messages%s", sdk.httpAdapterURL, chanID, subtopicPart)

	_, _, err := sdk.processRequest(http.MethodPost, url, ThingPrefix+key, []byte(msg), nil, http.StatusAccepted)

	return err
}

func (sdk mgSDK) ReadMessages(pm MessagePageMeta, chanName, token string) (MessagesPage, errors.SDKError) {
	chanNameParts := strings.SplitN(chanName, ".", channelParts)
	chanID := chanNameParts[0]
	subtopicPart := ""
	if len(chanNameParts) == channelParts {
		subtopicPart = fmt.Sprintf("?subtopic=%s", chanNameParts[1])
	}

	readMessagesEndpoint := fmt.Sprintf("channels/%s/messages%s", chanID, subtopicPart)
	url, err := sdk.withMessageQueryParams(sdk.readerURL, readMessagesEndpoint, pm)
	if err != nil {
		return MessagesPage{}, errors.NewSDKError(err)
	}

	fmt.Println("sdk url: ", url)

	header := make(map[string]string)
	header["Content-Type"] = string(sdk.msgContentType)

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, nil, header, http.StatusOK)
	if sdkerr != nil {
		return MessagesPage{}, sdkerr
	}

	var mp MessagesPage
	if err := json.Unmarshal(body, &mp); err != nil {
		return MessagesPage{}, errors.NewSDKError(err)
	}

	return mp, nil
}

func (sdk *mgSDK) SetContentType(ct ContentType) errors.SDKError {
	if ct != CTJSON && ct != CTJSONSenML && ct != CTBinary {
		return errors.NewSDKError(apiutil.ErrUnsupportedContentType)
	}

	sdk.msgContentType = ct

	return nil
}

func (sdk mgSDK) withMessageQueryParams(baseURL, endpoint string, mpm MessagePageMeta) (string, error) {
	q, err := mpm.mquery()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s?%s", baseURL, endpoint, q), nil
}

func (mpm MessagePageMeta) mquery() (string, error) {
	q := url.Values{}
	if mpm.Offset != 0 {
		q.Add("offset", strconv.FormatUint(mpm.Offset, 10))
	}
	if mpm.Limit != 0 {
		q.Add("limit", strconv.FormatUint(mpm.Limit, 10))
	}
	if mpm.Total != 0 {
		q.Add("total", strconv.FormatUint(mpm.Total, 10))
	}
	if mpm.Order != "" {
		q.Add("order", mpm.Order)
	}
	if mpm.Direction != "" {
		q.Add("dir", mpm.Direction)
	}
	if mpm.Level != 0 {
		q.Add("level", strconv.FormatUint(mpm.Level, 10))
	}
	if mpm.Identity != "" {
		q.Add("identity", mpm.Identity)
	}
	if mpm.Name != "" {
		q.Add("name", mpm.Name)
	}
	if mpm.Type != "" {
		q.Add("type", mpm.Type)
	}
	if mpm.Visibility != "" {
		q.Add("visibility", mpm.Visibility)
	}
	if mpm.Status != "" {
		q.Add("status", mpm.Status)
	}
	if mpm.Metadata != nil {
		md, err := json.Marshal(mpm.Metadata)
		if err != nil {
			return "", errors.NewSDKError(err)
		}
		q.Add("metadata", string(md))
	}
	if mpm.Action != "" {
		q.Add("action", mpm.Action)
	}
	if mpm.Subject != "" {
		q.Add("subject", mpm.Subject)
	}
	if mpm.Object != "" {
		q.Add("object", mpm.Object)
	}
	if mpm.Tag != "" {
		q.Add("tag", mpm.Tag)
	}
	if mpm.Owner != "" {
		q.Add("owner", mpm.Owner)
	}
	if mpm.SharedBy != "" {
		q.Add("shared_by", mpm.SharedBy)
	}
	if mpm.Topic != "" {
		q.Add("topic", mpm.Topic)
	}
	if mpm.Contact != "" {
		q.Add("contact", mpm.Contact)
	}
	if mpm.State != "" {
		q.Add("state", mpm.State)
	}
	if mpm.Permission != "" {
		q.Add("permission", mpm.Permission)
	}
	if mpm.ListPermissions != "" {
		q.Add("list_perms", mpm.ListPermissions)
	}
	if mpm.InvitedBy != "" {
		q.Add("invited_by", mpm.InvitedBy)
	}
	if mpm.UserID != "" {
		q.Add("user_id", mpm.UserID)
	}
	if mpm.DomainID != "" {
		q.Add("domain_id", mpm.DomainID)
	}
	if mpm.Relation != "" {
		q.Add("relation", mpm.Relation)
	}
	if mpm.Subtopic != "" {
		q.Add("subtopic", mpm.Subtopic)
	}
	if mpm.Publisher != "" {
		q.Add("publisher", mpm.Publisher)
	}
	if mpm.Comparator != "" {
		q.Add("comparator", mpm.Comparator)
	}
	if mpm.BoolValue {
		q.Add("bool_value", strconv.FormatBool(mpm.BoolValue))
	}
	if mpm.StringValue != "" {
		q.Add("string_value", mpm.StringValue)
	}
	if mpm.DataValue != "" {
		q.Add("data_value", mpm.DataValue)
	}
	if mpm.From != 0 {
		q.Add("from", strconv.FormatFloat(mpm.From, 'f', -1, 64))
	}
	if mpm.To != 0 {
		q.Add("to", strconv.FormatFloat(mpm.To, 'f', -1, 64))
	}
	if mpm.Total != 0 {
		q.Add("total", strconv.FormatUint(mpm.Total, 10))
	}
	if mpm.Aggregation != "" {
		q.Add("aggregation", mpm.Aggregation)
	}
	if mpm.Interval != "" {
		q.Add("interval", mpm.Interval)
	}
	return q.Encode(), nil
}
