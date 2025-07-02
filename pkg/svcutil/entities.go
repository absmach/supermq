package svcutil

import (
	"os"

	"gopkg.in/yaml.v2"
)

type EntityOperations struct {
	Operations         map[string]Permission `yaml:"operations"`
	RolesOperations    map[string]Permission `yaml:"roles_operations"`
	ExternalOperations map[string]Permission `yaml:"external_operations"`
}

type superMQEntities struct {
	Clients  EntityOperations `yaml:"clients"`
	Channels EntityOperations `yaml:"channels"`
	Groups   EntityOperations `yaml:"groups"`
	Domains  EntityOperations `yaml:"domains"`
	Users    EntityOperations `yaml:"users"`

	// Additional entities beyond the SuperMQ
	Additional map[string]EntityOperations `yaml:",inline"`
}

type EntitiesOperations interface {
	GetClients() EntityOperations
	GetChannels() EntityOperations
	GetGroups() EntityOperations
	GetDomains() EntityOperations
	GetUsers() EntityOperations
	GetAdditionalEntity(name string) (EntityOperations, bool)
}

func NewEntitiesOperationsFromPath(path string) (EntitiesOperations, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return NewEntitiesOperations(data)
}

func NewEntitiesOperations(data []byte) (EntitiesOperations, error) {
	var smqes superMQEntities
	if err := yaml.Unmarshal(data, &smqes); err != nil {
		return superMQEntities{}, err
	}

	return smqes, nil
}

func (s superMQEntities) GetClients() EntityOperations {
	return s.Clients
}

func (s superMQEntities) GetChannels() EntityOperations {
	return s.Channels
}

func (s superMQEntities) GetGroups() EntityOperations {
	return s.Groups
}

func (s superMQEntities) GetDomains() EntityOperations {
	return s.Domains
}

func (s superMQEntities) GetUsers() EntityOperations {
	return s.Users
}

func (s superMQEntities) GetAdditionalEntity(name string) (EntityOperations, bool) {
	eops, ok := s.Additional[name]
	return eops, ok
}
