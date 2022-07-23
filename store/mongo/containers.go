package mongo

import (
	"time"

	"github.com/FideTechSolutions/uberContainer/models"
	"github.com/FideTechSolutions/uberContainer/store"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (m *mongoStore) GetContainers() ([]models.Container, error) {

	session := m.Session.Clone()
	defer session.Close()

	c := session.DB(m.DatabaseName).C("containers")

	result := []models.Container{}
	err := c.Find(bson.M{}).All(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (m *mongoStore) GetContainer(key string) (models.Container, error) {

	session := m.Session.Clone()
	defer session.Close()

	c := session.DB(m.DatabaseName).C("containers")

	result := models.Container{}
	err := c.Find(bson.M{"key": key}).One(&result)
	if err != nil {
		if err == mgo.ErrNotFound {
			return models.Container{}, store.ErrNotFound
		}

		return models.Container{}, err
	}

	return result, nil
}

func (m *mongoStore) CreateContainer(container *models.Container) error {
	session := m.Session.Clone()
	defer session.Close()

	c := session.DB(m.DatabaseName).C("containers")

	container.CreatedAt = time.Now()
	container.ModifiedAt = time.Now()

	if err := c.Insert(container); err != nil {
		return err
	}

	return nil
}

func (m *mongoStore) UpdateContainer(container *models.Container) error {
	session := m.Session.Clone()
	defer session.Close()

	c := session.DB(m.DatabaseName).C("containers")

	container.ModifiedAt = time.Now()

	query := bson.M{"key": container.Key}

	err := c.Update(query, bson.M{
		"$set": container,
	})

	if err != nil {
		return err
	}

	return nil
}

func (m *mongoStore) DeleteContainer(key string) error {
	session := m.Session.Clone()
	defer session.Close()

	c := session.DB(m.DatabaseName).C("containers")

	err := c.Remove(bson.M{"key": key})
	if err != nil {
		return err
	}

	return err
}
