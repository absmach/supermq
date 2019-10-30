// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package transformer

import "github.com/mainflux/mainflux"

// Transformer specifies API form Message transformer.
type Transformer interface {
	// Transform Mainflux message to any other format.
	Transform(mainflux.Message) (interface{}, error)
}
