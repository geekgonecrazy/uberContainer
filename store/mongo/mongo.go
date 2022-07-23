package mongo

import (
	"github.com/FideTechSolutions/uberContainer/store"
	"gopkg.in/mgo.v2"
)

type mongoStore struct {
	DatabaseName string
	*mgo.Session
}

func New(host string) (store.Store, error) {
	dbName, sess, err := connect(host)
	if err != nil {
		return nil, err
	}

	s := &mongoStore{dbName, sess}

	s.EnsureIndexes()

	return s, nil
}

func connect(connectionstring string) (string, *mgo.Session, error) {

	dailInfo, err := mgo.ParseURL(connectionstring)
	if err != nil {
		return "", nil, err
	}

	sess, err := mgo.DialWithInfo(dailInfo)
	if err != nil {
		return "", nil, err
	}

	return dailInfo.Database, sess.Copy(), nil
}

func (m *mongoStore) CheckDb() error {
	sess := m.Session.Copy()
	defer sess.Close()

	if err := sess.Ping(); err != nil {
		return err
	}

	return nil
}

// EnsureIndexes ensures the indexes are in place
func (m *mongoStore) EnsureIndexes() error {
	sess := m.Session.Copy()
	defer sess.Close()

	if err := sess.DB(m.DatabaseName).C("containers").EnsureIndex(mgo.Index{Key: []string{"key"}, Unique: true}); err != nil {
		return err
	}

	return nil
}
