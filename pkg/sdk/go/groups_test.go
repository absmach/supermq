package sdk_test

import (
	"fmt"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	group = sdk.Group{
		ID:          "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Name:        "test1",
		Description: "desc",
		ParentID:    "parentId",
		Metadata:    nil,
	}
)

func TestCreateGroup(t *testing.T) {
	svc := newThingsService(map[string]string{token: email})
	as := newAuthServer(svc)

	defer as.Close()

	sdkConf := sdk.Config{
		AuthURL:         as.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc  string
		group sdk.Group
		token string
		err   error
		empty bool
	}{
		{
			desc:  "create new group",
			group: group,
			token: token,
			err:   nil,
			empty: true,
		},
	}
	for _, tc := range cases {
		loc, err := mainfluxSDK.CreateGroup(tc.group, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		assert.Equal(t, tc.empty, loc == "", fmt.Sprintf("%s: expected empty result location, got: %s", tc.desc, loc))
	}
}
