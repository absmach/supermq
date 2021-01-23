// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"encoding/json"
	"sync"

	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
)

var _ readers.MessageRepository = (*messageRepositoryMock)(nil)

type messageRepositoryMock struct {
	mutex    sync.Mutex
	messages map[string][]readers.Message
}

// NewMessageRepository returns mock implementation of message repository.
func NewMessageRepository(chanID string, messages []readers.Message) readers.MessageRepository {
	repo := map[string][]readers.Message{
		chanID: messages,
	}

	return &messageRepositoryMock{
		mutex:    sync.Mutex{},
		messages: repo,
	}
}

func (repo *messageRepositoryMock) ReadAll(chanID string, rpm readers.PageMetadata) (readers.MessagesPage, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if rpm.Format != "" && rpm.Format != "messages" {
		return readers.MessagesPage{}, nil
	}

	var query map[string]interface{}
	meta, _ := json.Marshal(rpm)
	json.Unmarshal(meta, &query)

	var msgs []readers.Message
	filter := false
	for _, m := range repo.messages[chanID] {
		senml := m.(senml.Message)
		filterOk := false

	forLoop:
		for name := range query {
			switch name {
			case "subtopic":
				filter, filterOk = true, false
				if rpm.Subtopic == senml.Subtopic {
					filterOk = true
				} else {
					break forLoop
				}
			case "publisher":
				filter, filterOk = true, false
				if rpm.Publisher == senml.Publisher {
					filterOk = true
				} else {
					break forLoop
				}
			case "name":
				filter, filterOk = true, false
				if rpm.Name == senml.Name {
					filterOk = true
				} else {
					break forLoop
				}
			case "protocol":
				filter, filterOk = true, false
				if rpm.Protocol == senml.Protocol {
					filterOk = true
				} else {
					break forLoop
				}
			case "v":
				filter, filterOk = true, false
				if senml.Value != nil &&
					*senml.Value == rpm.Value {
					filterOk = true
				} else {
					break forLoop
				}
			case "vb":
				filter, filterOk = true, false
				if senml.BoolValue != nil &&
					*senml.BoolValue == rpm.BoolValue {
					filterOk = true
				} else {
					break forLoop
				}
			case "vs":
				filter, filterOk = true, false
				if senml.StringValue != nil &&
					*senml.StringValue == rpm.StringValue {
					filterOk = true
				} else {
					break forLoop
				}
			case "vd":
				filter, filterOk = true, false
				if senml.DataValue != nil &&
					*senml.DataValue == rpm.DataValue {
					filterOk = true
				} else {
					break forLoop
				}
			case "from":
				filter, filterOk = true, false
				if senml.Time >= rpm.From {
					filterOk = true
				} else {
					break forLoop
				}
			case "to":
				filter, filterOk = true, false
				if senml.Time < rpm.To {
					filterOk = true
				} else {
					break forLoop
				}
			}
		}

		if filter && filterOk {
			msgs = append(msgs, m)
		}

		if !filter {
			msgs = append(msgs, m)
		}
	}

	numOfMessages := uint64(len(msgs))

	if rpm.Offset >= numOfMessages {
		return readers.MessagesPage{}, nil
	}

	if rpm.Limit < 1 {
		return readers.MessagesPage{}, nil
	}

	end := rpm.Offset + rpm.Limit
	if rpm.Offset+rpm.Limit > numOfMessages {
		end = numOfMessages
	}

	return readers.MessagesPage{
		PageMetadata: rpm,
		Total:        uint64(len(msgs)),
		Messages:     msgs[rpm.Offset:end],
	}, nil
}
