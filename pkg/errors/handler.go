// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

type Mapper interface {
	GetError(key string) (error, bool)
}

type Handler interface {
	HandleError(wrapper, err error) error
}

type HandlerOption func(*Handler)
