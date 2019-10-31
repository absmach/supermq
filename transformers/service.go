// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package transformers

import "github.com/mainflux/mainflux"

// Service specifies API form Message transformer.
type Service interface {
	// Transform Mainflux message to any other format.
	Transform(mainflux.Message) (interface{}, error)
}
