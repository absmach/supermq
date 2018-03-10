package postgres

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mainflux/mainflux/manager"
)

var _ manager.ChannelRepository = (*channelRepository)(nil)

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) manager.ChannelRepository {
	return &channelRepository{db}
}

func (cr channelRepository) Save(channel manager.Channel) (string, error) {
	return "", nil
}

func (cr channelRepository) Update(channel manager.Channel) error {
	return nil
}

func (cr channelRepository) One(owner, id string) (manager.Channel, error) {
	return manager.Channel{}, manager.ErrNotFound
}

func (cr channelRepository) All(owner string) []manager.Channel {
	return make([]manager.Channel, 0)
}

func (cr channelRepository) Remove(owner, id string) error {
	return nil
}

func (cr channelRepository) Connect(owner, channel, client string) error {
	return nil
}

func (cr channelRepository) Disconnect(owner, channel, client string) error {
	return nil
}

func (cr channelRepository) HasClient(channel, client string) bool {
	return false
}
