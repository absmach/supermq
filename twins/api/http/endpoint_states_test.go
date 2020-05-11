// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/mocks"
	"github.com/stretchr/testify/require"
)

func TestListStates(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	sr := mocks.NewStateRepository()
	defer ts.Close()

	twin := twins.Twin{
		Owner: email,
	}
	def := twins.Definition{}
	stw, _ := svc.AddTwin(context.Background(), token, twin, def)

	data := []stateRes{}
	for i := 0; i < 100; i++ {
		st := twins.State{
			TwinID:     stw.ID,
			ID:         int64(i),
			Definition: 0,
			Created:    time.Now().Round(0),
		}
		err := sr.Save(context.TODO(), st)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		stres := stateRes{
			TwinID:     st.TwinID,
			ID:         st.ID,
			Definition: st.Definition,
			Created:    st.Created,
		}
		data = append(data, stres)
	}
}

type stateRes struct {
	TwinID     string                 `json:"twin_id"`
	ID         int64                  `json:"id"`
	Definition int                    `json:"definition"`
	Created    time.Time              `json:"created"`
	Payload    map[string]interface{} `json:"payload"`
}

type statesPageRes struct {
	pageRes
	States []stateRes `json:"states"`
}
