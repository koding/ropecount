package mongodb

import (
	"gopkg.in/mgo.v2"
)

// MongoDB holds the contextual info for mongo db session
type MongoDB struct {
	Session *mgo.Session
	URL     string
}

// New creates a new mongo db connection
func New(url string) (*MongoDB, error) {
	m := &MongoDB{
		URL: url,
	}

	mgo.SetStats(true)

	var err error
	if m.Session, err = mgo.Dial(url); err != nil {
		return nil, err
	}

	m.Session.SetSafe(&mgo.Safe{})
	m.Session.SetMode(mgo.Strong, true)
	return m, nil
}

// Close closes the db connection
func (m *MongoDB) Close() {
	m.Session.Close()
}

// Copy creates a thread safe copy of the connection
func (m *MongoDB) Copy() *mgo.Session {
	return m.Session.Copy()
}

// Run gets the collection from the db and runs the given function.
func (m *MongoDB) Run(collection string, s func(*mgo.Collection) error) error {
	session := m.Copy()
	defer session.Close()
	c := session.DB("").C(collection)
	return s(c)
}
