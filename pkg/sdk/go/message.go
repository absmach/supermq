// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

func (sdk mfSDK) SendMessage(chanName, msg, key string) errors.SDKError {
	chanNameParts := strings.SplitN(chanName, ".", 2)
	chanID := chanNameParts[0]
	subtopicPart := ""
	if len(chanNameParts) == 2 {
		subtopicPart = fmt.Sprintf("/%s", strings.Replace(chanNameParts[1], ".", "/", -1))
	}

	url := fmt.Sprintf("%s/channels/%s/messages/%s", sdk.httpAdapterURL, chanID, subtopicPart)

	_, _, err := sdk.processRequest(http.MethodPost, url, []byte(msg), apiutil.ThingPrefix+key, string(CTJSON), http.StatusAccepted)
	return err
	// req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(msg))
	// if err != nil {
	// 	return errors.NewSDKError(err)
	// }

	// resp, err := sdk.sendThingRequest(req, key, string(sdk.msgContentType))
	// if err != nil {
	// 	return errors.NewSDKError(err)
	// }
	// defer resp.Body.Close()

	// if err := errors.CheckError(resp, http.StatusAccepted); err != nil {
	// 	return err
	// }

	// return nil
}

func (sdk mfSDK) ReadMessages(chanName, token string) (MessagesPage, errors.SDKError) {
	chanNameParts := strings.SplitN(chanName, ".", 2)
	chanID := chanNameParts[0]
	subtopicPart := ""
	if len(chanNameParts) == 2 {
		subtopicPart = fmt.Sprintf("?subtopic=%s", strings.Replace(chanNameParts[1], ".", "/", -1))
	}

	url := fmt.Sprintf("%s/channels/%s/messages%s", sdk.readerURL, chanID, subtopicPart)

	body, _, err := sdk.processRequest(http.MethodGet, url, nil, token, string(sdk.msgContentType), http.StatusOK)
	if err != nil {
		return MessagesPage{}, err
	}

	var mp MessagesPage
	if err := json.Unmarshal(body, &mp); err != nil {
		return MessagesPage{}, errors.NewSDKError(err)
	}

	return mp, nil
}

func (sdk *mfSDK) SetContentType(ct ContentType) errors.SDKError {
	if ct != CTJSON && ct != CTJSONSenML && ct != CTBinary {
		return errors.NewSDKError(errors.ErrUnsupportedContentType)
	}

	sdk.msgContentType = ct
	return nil
}
