package mongodb

import (
	"gopkg.in/mgo.v2"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/writers"
)

var _ writers.MessageRepository = (*mongoRepo)(nil)

type mongoRepo struct {
	dbName         string
	collectionName string
	db             *mgo.Session
}

// New returns new MongoDB writer.
func New(dbName string, collectionName string, db *mgo.Session) (writers.MessageRepository, error) {
	return &mongoRepo{dbName, collectionName, db}, nil
}

func (repo *mongoRepo) Save(msg mainflux.Message) error {
	s := repo.db.Copy()
	defer s.Close()
	c := s.DB(repo.dbName).C(repo.collectionName)

	return c.Insert(msg)
}
