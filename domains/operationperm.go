// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package domains

import "fmt"

type Operation int

func (op Operation) String() string {
	switch op {
	case OpUpdateDomain:
		return OpUpdateDomainStr
	case OpRetrieveDomain:
		return OpRetrieveDomainStr
	case OpEnableDomain:
		return OpEnableDomainStr
	case OpDisableDomain:
		return OpDisableDomainStr
	case OpSendInvitation:
		return OpSendInvitationStr
	case OpAcceptInvitation:
		return OpAcceptInvitationStr
	case OpCreateDomain:
		return OpCreateDomainStr
	case OpFreezeDomain:
		return OpFreezeDomainStr
	case OpListDomains:
		return OpListDomainsStr
	case OpViewInvitation:
		return OpViewInvitationStr
	case OpListInvitations:
		return OpListInvitationsStr
	case OpRejectInvitation:
		return OpRejectInvitationStr
	case OpDeleteInvitation:
		return OpDeleteInvitationStr
	default:
		return fmt.Sprintf("unknown operation: %d", op)
	}
}

type OperationPerm struct {
	opPerm      map[Operation]Permission
	expectedOps []Operation
}

func newOperationPerm(expectedOps []Operation) OperationPerm {
	return OperationPerm{
		opPerm:      make(map[Operation]Permission),
		expectedOps: expectedOps,
	}
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

func (opp OperationPerm) isKeyRequired(op Operation) bool {
	for _, key := range opp.expectedOps {
		if key == op {
			return true
		}
	}
	return false
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
