// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package channels

import "fmt"

type Operation int

func (op Operation) String() string {
	switch op {
	case OpViewChannel:
		return OpViewChannelStr
	case OpUpdateChannel:
		return OpUpdateChannelStr
	case OpUpdateChannelTags:
		return OpUpdateChannelTagsStr
	case OpEnableChannel:
		return OpEnableChannelStr
	case OpDisableChannel:
		return OpDisableChannelStr
	case OpDeleteChannel:
		return OpDeleteChannelStr
	case OpSetParentGroup:
		return OpSetParentGroupStr
	case OpRemoveParentGroup:
		return OpRemoveParentGroupStr
	case OpConnectClient:
		return OpConnectClientStr
	case OpDisconnectClient:
		return OpDisconnectClientStr
	case OpCreateChannel:
		return OpCreateChannelStr
	case OpListChannels:
		return OpListChannelsStr
	case OpListUserChannels:
		return OpListUserChannelsStr
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

type ExternalOperation int

func (ep ExternalOperation) String() string {
	switch ep {
	case DomainOpCreateChannel:
		return DomainOpCreateChannelStr
	case DomainOpListChannel:
		return DomainOpListChannelStr
	case GroupOpSetChildChannel:
		return GroupOpSetChildChannelStr
	case GroupsOpRemoveChildChannel:
		return GroupsOpRemoveChildChannelStr
	case ClientsOpConnectChannel:
		return ClientsOpConnectChannelStr
	case ClientsOpDisconnectChannel:
		return ClientsOpDisconnectChannelStr
	default:
		return fmt.Sprintf("unknown external operation: %d", ep)
	}
}

type ExternalOperationPerm struct {
	opPerm      map[ExternalOperation]Permission
	expectedOps []ExternalOperation
}

func newExternalOperationPerm(expectedOps []ExternalOperation) ExternalOperationPerm {
	return ExternalOperationPerm{
		opPerm:      make(map[ExternalOperation]Permission),
		expectedOps: expectedOps,
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
			return fmt.Errorf("%v is not a valid external operation", eop.String())
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
			return fmt.Errorf("ExternalOperationPerm: \"%s\" is not a valid external operation", eop.String())
		}
	}
	for _, eeo := range eopp.expectedOps {
		if _, ok := eopp.opPerm[eeo]; !ok {
			return fmt.Errorf("ExternalOperationPerm: \"%s\" external operation is missing", eeo.String())
		}
	}
	return nil
}

func (eopp ExternalOperationPerm) GetPermission(eop ExternalOperation) (Permission, error) {
	if perm, ok := eopp.opPerm[eop]; ok {
		return perm, nil
	}
	return "", fmt.Errorf("external operation \"%s\" doesn't have any permissions", eop.String())
}
