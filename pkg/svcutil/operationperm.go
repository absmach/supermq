// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package svcutil

import (
	"fmt"
)

type Permission string

func (p Permission) String() string {
	return string(p)
}

type Operation int

type ExternalOperation int

type RoleOperation int

type OperationKey interface {
	Operation | ExternalOperation | RoleOperation
}

type OperationDetails struct {
	Name               string
	PermissionRequired bool
}

type operations[K OperationKey] struct {
	opp map[K]Permission
	opd map[K]OperationDetails
}

type Operations[K OperationKey] interface {
	GetPermission(op K) (Permission, error)
	GetPermissionAndRequired(op K) (Permission, bool, error)
	OperationName(op K) string
	Validate() error
}

func NewOperations[K OperationKey](opd map[K]OperationDetails, opp map[string]Permission) (Operations[K], error) {
	ops := newEmptyOperations(opd)

	if err := ops.addOperationPermission(opp); err != nil {
		return nil, err
	}
	if err := ops.Validate(); err != nil {
		return nil, err
	}
	return &ops, nil
}

func newEmptyOperations[K OperationKey](opd map[K]OperationDetails) operations[K] {
	return operations[K]{
		opp: make(map[K]Permission),
		opd: opd,
	}
}

func (ops *operations[K]) OperationName(op K) string {
	opd, ok := ops.opd[op]
	if !ok {
		return fmt.Sprintf("UnknownOperation(%v)", op)
	}
	return opd.Name
}

func (ops *operations[K]) addOperationPermission(opnamePerm map[string]Permission) error {
	for op, opd := range ops.opd {
		if opd.PermissionRequired {
			perm, ok := opnamePerm[opd.Name]
			if !ok {
				return fmt.Errorf("permission related to operation name %s not found", opd.Name)
			}
			ops.opp[op] = perm
		}
	}
	return nil
}

func (ops *operations[K]) Validate() error {
	for op, opd := range ops.opd {
		if opd.PermissionRequired {
			if _, ok := ops.opp[op]; !ok {
				return fmt.Errorf("permission related to operation name %s not found", opd.Name)
			}
		}
	}
	return nil
}

func (ops *operations[K]) GetPermission(op K) (Permission, error) {
	if perm, ok := ops.opp[op]; ok {
		return perm, nil
	}
	return "", fmt.Errorf("operation %s doesn't have any permissions", ops.OperationName(op))
}

func (ops *operations[K]) GetPermissionAndRequired(op K) (Permission, bool, error) {
	opd, ok := ops.opd[op]
	if !ok {
		return "", false, fmt.Errorf("%s", ops.OperationName(op))
	}
	perm, ok := ops.opp[op]
	if opd.PermissionRequired && !ok {
		return "", false, fmt.Errorf("operation %s doesn't have any permissions", ops.OperationName(op))
	}
	return perm, opd.PermissionRequired, nil
}
