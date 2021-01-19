// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"sync"

	"github.com/mainflux/mainflux/readers"
)

var _ readers.MessageRepository = (*messageRepositoryMock)(nil)

type messageRepositoryMock struct {
	mutex    sync.Mutex
	messages map[string][]readers.Message
}

// NewMessageRepository returns mock implementation of message repository.
func NewMessageRepository(messages map[string][]readers.Message) readers.MessageRepository {
	return &messageRepositoryMock{
		mutex:    sync.Mutex{},
		messages: messages,
	}
}

func (repo *messageRepositoryMock) ReadAll(chanID string, rpm readers.PageMetadata) (readers.MessagesPage, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	end := rpm.Offset + rpm.Limit

	numOfMessages := uint64(len(repo.messages[chanID]))
	if rpm.Offset < 0 || rpm.Offset >= numOfMessages {
		return readers.MessagesPage{}, nil
	}

	if rpm.Limit < 1 {
		return readers.MessagesPage{}, nil
	}

	if rpm.Offset+rpm.Limit > numOfMessages {
		end = numOfMessages
	}

	return readers.MessagesPage{
		PageMetadata: rpm,
		Messages:     repo.messages[chanID][rpm.Offset:end],
	}, nil
}
