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
	for _, m := range repo.messages[chanID] {
		senml := m.(senml.Message)

		filter := false
		for name := range query {
			switch name {
			case "subtopic":
				filter = true
				if rpm.Subtopic == senml.Subtopic {
					msgs = append(msgs, m)
				}
			case "publisher":
				filter = true
				if rpm.Publisher == senml.Publisher {
					msgs = append(msgs, m)
				}
			case "name":
				filter = true
				if rpm.Name == senml.Name {
					msgs = append(msgs, m)
				}
			case "protocol":
				filter = true
				if rpm.Protocol == senml.Protocol {
					msgs = append(msgs, m)
				}
			case "v":
				filter = true
				if senml.Value != nil &&
					*senml.Value == rpm.Value {
					msgs = append(msgs, m)
				}
			case "vb":
				filter = true
				if senml.BoolValue != nil &&
					*senml.BoolValue == rpm.BoolValue {
					msgs = append(msgs, m)
				}
			case "vs":
				filter = true
				if senml.StringValue != nil &&
					*senml.StringValue == rpm.StringValue {
					msgs = append(msgs, m)
				}
			case "vd":
				filter = true
				if senml.DataValue != nil &&
					*senml.DataValue == rpm.DataValue {
					msgs = append(msgs, m)
				}
			case "from":
				filter = true
				if senml.Time >= rpm.From {
					msgs = append(msgs, m)
				}
			case "to":
				filter = true
				if senml.Time < rpm.To {
					msgs = append(msgs, m)
				}
			}
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
		Messages:     msgs[rpm.Offset:end],
	}, nil
}
