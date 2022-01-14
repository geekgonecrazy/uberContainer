package mongo

import (
	"github.com/geekgonecrazy/uberContainer/models"
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
	err := c.Find(bson.M{"_id": key}).One(&result)
	if err != nil {
		return models.Container{}, err
	}

	return result, nil
}

func (m *mongoStore) CreateContainer(container *models.Container) error {
	session := m.Session.Clone()
	defer session.Close()

	c := session.DB(m.DatabaseName).C("containers")

	if err := c.Insert(container); err != nil {
		return err
	}

	return nil
}

func (m *mongoStore) UpdateContainer(container *models.Container) error {
	session := m.Session.Clone()
	defer session.Close()

	c := session.DB(m.DatabaseName).C("containers")

	query := bson.M{"_id": container.Key}

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

	err := c.RemoveId(key)
	if err != nil {
		return err
	}

	return err
}
