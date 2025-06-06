// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package svcutil

import "fmt"

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

func NewExternalOperationPerm(expectedOps []ExternalOperation, opNames []string) ExternalOperationPerm {
	return ExternalOperationPerm{
		opPerm:      make(map[ExternalOperation]Permission),
		expectedOps: expectedOps,
		opNames:     opNames,
	}
}

func (req ExternalOperationPerm) isKeyRequired(eop ExternalOperation) bool {
	for _, key := range req.expectedOps {
		if key == eop {
			return true
		}
	}
	return false
}

func (req ExternalOperationPerm) AddOperationPermissionMap(eopMap map[ExternalOperation]Permission) error {
	// First iteration check all the keys are valid, If any one key is invalid then no key should be added.
	for eop := range eopMap {
		if !req.isKeyRequired(eop) {
			return fmt.Errorf("%v is not a valid external operation", eop.String(req.opNames))
		}
	}
	for eop, perm := range eopMap {
		req.opPerm[eop] = perm
	}
	return nil
}

func (req ExternalOperationPerm) AddOperationPermission(eop ExternalOperation, perm Permission) error {
	if !req.isKeyRequired(eop) {
		return fmt.Errorf("%v is not a valid external operation", eop.String(req.opNames))
	}
	req.opPerm[eop] = perm
	return nil
}

func (req ExternalOperationPerm) Validate() error {
	for eop := range req.opPerm {
		if !req.isKeyRequired(eop) {
			return fmt.Errorf("ExternalOperationPerm: \"%s\" is not a valid external operation", eop.String(req.opNames))
		}
	}
	for _, eeo := range req.expectedOps {
		if _, ok := req.opPerm[eeo]; !ok {
			return fmt.Errorf("ExternalOperationPerm: \"%s\" external operation is missing", eeo.String(req.opNames))
		}
	}
	return nil
}

func (req ExternalOperationPerm) GetPermission(eop ExternalOperation) (Permission, error) {
	if perm, ok := req.opPerm[eop]; ok {
		return perm, nil
	}
	return "", fmt.Errorf("external operation \"%s\" doesn't have any permissions", eop.String(req.opNames))
}
