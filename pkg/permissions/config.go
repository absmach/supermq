// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package permissions

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type PermissionConfig struct {
	Entities map[string]EntityPermissions `yaml:",inline"`
}

type EntityPermissions struct {
	Operations      []map[string]string `yaml:"operations"`
	RolesOperations []map[string]string `yaml:"roles_operations"`
}

func ParsePermissionsFile(filePath string) (*PermissionConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read permissions file: %w", err)
	}

	var config PermissionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse permissions file: %w", err)
	}

	return &config, nil
}

func (pc *PermissionConfig) GetEntityPermissions(entityType string) (map[string]Permission, map[string]Permission, error) {
	entityPerms, ok := pc.Entities[entityType]
	if !ok {
		return nil, nil, fmt.Errorf("entity type %s not found in permissions file", entityType)
	}

	operations := make(map[string]Permission)
	for _, op := range entityPerms.Operations {
		for name, perm := range op {
			operations[name] = Permission(perm)
		}
	}

	rolesOperations := make(map[string]Permission)
	for _, op := range entityPerms.RolesOperations {
		for name, perm := range op {
			rolesOperations[name] = Permission(perm)
		}
	}

	return operations, rolesOperations, nil
}
