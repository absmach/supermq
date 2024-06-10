// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	mgxsdk "github.com/absmach/magistrala/pkg/sdk/go"
)

// Keep SDK handle in global var.
var sdk mgxsdk.SDK

// SetSDK sets magistrala SDK instance.
func SetSDK(s mgxsdk.SDK) {
	sdk = s
}

func AccessCurlFlagChan() {
	curlFlagChan := sdk.GetCurlFlagChan()
	go func() {
		for curlCommand := range curlFlagChan {
			fmt.Println(curlCommand)
		}
	}()
}
