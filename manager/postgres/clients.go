package postgres

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mainflux/mainflux/manager"
	uuid "github.com/satori/go.uuid"
)

var _ manager.ClientRepository = (*clientRepository)(nil)

type clientRepository struct {
	db *gorm.DB
}

// NewClientRepository instantiates a PostgreSQL implementation of client
// repository.
func NewClientRepository(db *gorm.DB) manager.ClientRepository {
	return &clientRepository{db}
}

func (cr *clientRepository) Id() string {
	return uuid.NewV4().String()
}

func (cr *clientRepository) Save(client manager.Client) error {
	if err := cr.db.Create(&client).Error; err != nil {
		return err
	}

	return nil
}

func (cr *clientRepository) Update(client manager.Client) error {
	// This unfortunate extra query is introduced due to the fact that updating
	// a non-existent entry does not return an error. If at some point becomes
	// possible to retrieve an error from update, it should be removed.
	if _, err := cr.One(client.Owner, client.ID); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return manager.ErrNotFound
		}

		return err
	}

	return cr.db.Model(&client).Updates(client).Error
}

func (cr *clientRepository) One(owner, id string) (manager.Client, error) {
	client := manager.Client{}

	q := cr.db.Where("owner = ? AND id = ?", owner, id)

	if err := q.First(&client).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return client, manager.ErrNotFound
		}

		return client, err
	}

	return client, nil
}

func (cr *clientRepository) All(owner string) []manager.Client {
	var clients []manager.Client

	cr.db.Where("owner = ?", owner).Find(&clients)

	return clients
}

func (cr *clientRepository) Remove(owner, id string) error {
	q := cr.db.Where("owner = ? AND id = ?", owner, id)
	q.Delete(manager.Client{})
	return nil
}
