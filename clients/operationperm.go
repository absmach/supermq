// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package clients

import "fmt"

type Operation int

func (op Operation) String() string {
	if (int(op) < 0) || (int(op) == len(OperationNames)) {
		return fmt.Sprintf("UnknownOperation(%d)", op)
	}
	return OperationNames[op]
}

type OperationPerm struct {
	opPerm      map[Operation]Permission
	expectedOps []Operation
	opNames     []string
}

func newOperationPerm(expectedOps []Operation, opNames []string) OperationPerm {
	return OperationPerm{
		opPerm:      make(map[Operation]Permission),
		expectedOps: expectedOps,
		opNames:     opNames,
	}
}

func (opp OperationPerm) isKeyRequired(op Operation) bool {
	for _, key := range opp.expectedOps {
		if key == op {
			return true
		}
	}
	return false
}

func (opp OperationPerm) AddOperationPermissionMap(opMap map[Operation]Permission) error {
	// First iteration check all the keys are valid, If any one key is invalid then no key should be added.
	for op := range opMap {
		if !opp.isKeyRequired(op) {
			return fmt.Errorf("%v is not a valid operation", op.String())
		}
	}
	for op, perm := range opMap {
		opp.opPerm[op] = perm
	}
	return nil
}

func (opp OperationPerm) AddOperationPermission(op Operation, perm Permission) error {
	if !opp.isKeyRequired(op) {
		return fmt.Errorf("%v is not a valid operation", op.String())
	}
	opp.opPerm[op] = perm
	return nil
}

func (opp OperationPerm) Validate() error {
	for op := range opp.opPerm {
		if !opp.isKeyRequired(op) {
			return fmt.Errorf("OperationPerm: \"%s\" is not a valid operation", op.String())
		}
	}
	for _, eeo := range opp.expectedOps {
		if _, ok := opp.opPerm[eeo]; !ok {
			return fmt.Errorf("OperationPerm: \"%s\" operation is missing", eeo.String())
		}
	}
	return nil
}

func (opp OperationPerm) GetPermission(op Operation) (Permission, error) {
	if perm, ok := opp.opPerm[op]; ok {
		return perm, nil
	}
	return "", fmt.Errorf("operation \"%s\" doesn't have any permissions", op.String())
}

type Permission string

func (p Permission) String() string {
	return string(p)
}

type ExternalOperation int

func (op ExternalOperation) String(operations []string) string {
	if (int(op) < 0) || (int(op) == len(operations)) {
		return fmt.Sprintf("UnknownOperation(%d)", op)
	}
	return operations[op]
}

type ExternalOperationPerm struct {
	opPerm      map[ExternalOperation]Permission
	expectedOps []ExternalOperation
	opNames     []string
}

func newExternalOperationPerm(expectedOps []ExternalOperation, opNames []string) ExternalOperationPerm {
	return ExternalOperationPerm{
		opPerm:      make(map[ExternalOperation]Permission),
		expectedOps: expectedOps,
		opNames:     opNames,
	}
}

func (eopp ExternalOperationPerm) isKeyRequired(eop ExternalOperation) bool {
	for _, key := range eopp.expectedOps {
		if key == eop {
			return true
		}
	}
	return false
}

func (eopp ExternalOperationPerm) AddOperationPermissionMap(eopMap map[ExternalOperation]Permission) error {
	// First iteration check all the keys are valid, If any one key is invalid then no key should be added.
	for eop := range eopMap {
		if !eopp.isKeyRequired(eop) {
			return fmt.Errorf("%v is not a valid external operation", eop.String(eopp.opNames))
		}
	}
	for eop, perm := range eopMap {
		eopp.opPerm[eop] = perm
	}
	return nil
}

func (eopp ExternalOperationPerm) Validate() error {
	for eop := range eopp.opPerm {
		if !eopp.isKeyRequired(eop) {
			return fmt.Errorf("ExternalOperationPerm: \"%s\" is not a valid external operation", eop.String(eopp.opNames))
		}
	}
	for _, eeo := range eopp.expectedOps {
		if _, ok := eopp.opPerm[eeo]; !ok {
			return fmt.Errorf("ExternalOperationPerm: \"%s\" external operation is missing", eeo.String(eopp.opNames))
		}
	}
	return nil
}

func (eopp ExternalOperationPerm) GetPermission(eop ExternalOperation) (Permission, error) {
	if perm, ok := eopp.opPerm[eop]; ok {
		return perm, nil
	}
	return "", fmt.Errorf("external operation \"%s\" doesn't have any permissions", eop.String(eopp.opNames))
}
