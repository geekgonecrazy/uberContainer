package store

import (
	"errors"

	"github.com/FideTechSolutions/uberContainer/models"
)

//Store is an interface that the store should implement
type Store interface {
	CreateContainer(container *models.Container) error
	GetContainers() ([]models.Container, error)
	GetContainer(key string) (models.Container, error)
	UpdateContainer(container *models.Container) error
	DeleteContainer(id string) error

	CheckDb() error
}

var ErrNotFound = errors.New("record not found")
