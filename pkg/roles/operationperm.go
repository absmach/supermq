// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package roles

import "fmt"

type Operation int

const (
	OpAddRole Operation = iota
	OpRemoveRole
	OpUpdateRoleName
	OpRetrieveRole
	OpRetrieveAllRoles
	OpRoleAddActions
	OpRoleListActions
	OpRoleCheckActionsExists
	OpRoleRemoveActions
	OpRoleRemoveAllActions
	OpRoleAddMembers
	OpRoleListMembers
	OpRoleCheckMembersExists
	OpRoleRemoveMembers
	OpRoleRemoveAllMembers
	OpListAvailableActions
)

var expectedOperations = []Operation{
	OpAddRole,
	OpRemoveRole,
	OpUpdateRoleName,
	OpRetrieveRole,
	OpRetrieveAllRoles,
	OpRoleAddActions,
	OpRoleListActions,
	OpRoleCheckActionsExists,
	OpRoleRemoveActions,
	OpRoleRemoveAllActions,
	OpRoleAddMembers,
	OpRoleListMembers,
	OpRoleCheckMembersExists,
	OpRoleRemoveMembers,
	OpRoleRemoveAllMembers,
}

const (
	OpAddRoleStr                = "OpAddRole"
	OpRemoveRoleStr             = "OpRemoveRole"
	OpUpdateRoleNameStr         = "OpUpdateRoleName"
	OpRetrieveRoleStr           = "OpRetrieveRole"
	OpRetrieveAllRolesStr       = "OpRetrieveAllRoles"
	OpRoleAddActionsStr         = "OpRoleAddActions"
	OpRoleListActionsStr        = "OpRoleListActions"
	OpRoleCheckActionsExistsStr = "OpRoleCheckActionsExists"
	OpRoleRemoveActionsStr      = "OpRoleRemoveActions"
	OpRoleRemoveAllActionsStr   = "OpRoleRemoveAllActions"
	OpRoleAddMembersStr         = "OpRoleAddMembers"
	OpRoleListMembersStr        = "OpRoleListMembers"
	OpRoleCheckMembersExistsStr = "OpRoleCheckMembersExists"
	OpRoleRemoveMembersStr      = "OpRoleRemoveMembers"
	OpRoleRemoveAllMembersStr   = "OpRoleRemoveAllMembers"
	OpListAvailableActionsStr   = "OpListAvailableActions"
)

func (op Operation) String() string {
	switch op {
	case OpAddRole:
		return OpAddRoleStr
	case OpRemoveRole:
		return OpRemoveRoleStr
	case OpUpdateRoleName:
		return OpUpdateRoleNameStr
	case OpRetrieveRole:
		return OpRetrieveRoleStr
	case OpRetrieveAllRoles:
		return OpRetrieveAllRolesStr
	case OpRoleAddActions:
		return OpRoleAddActionsStr
	case OpRoleListActions:
		return OpRoleListActionsStr
	case OpRoleCheckActionsExists:
		return OpRoleCheckActionsExistsStr
	case OpRoleRemoveActions:
		return OpRoleRemoveActionsStr
	case OpRoleRemoveAllActions:
		return OpRoleRemoveAllActionsStr
	case OpRoleAddMembers:
		return OpRoleAddMembersStr
	case OpRoleListMembers:
		return OpRoleListMembersStr
	case OpRoleCheckMembersExists:
		return OpRoleCheckMembersExistsStr
	case OpRoleRemoveMembers:
		return OpRoleRemoveMembersStr
	case OpRoleRemoveAllMembers:
		return OpRoleRemoveAllMembersStr
	case OpListAvailableActions:
		return OpListAvailableActionsStr
	default:
		return fmt.Sprintf("unknown operation: %d", op)
	}
}

type Permission string

func (p Permission) String() string {
	return string(p)
}

type OperationPerm struct {
	opPerm      map[Operation]Permission
	expectedOps []Operation
}

func NewOperationPerm() OperationPerm {
	return newOperationPerm(expectedOperations)
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
