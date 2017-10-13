package mongodb

import (
	"errors"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Compaction holds the parts of a key as separate entities
type Compaction struct {
	ID        bson.ObjectId    `bson:"_id,omitempty" json:"_id"`
	Direction string           `bson:"direction" json:"direction"`
	Segment   string           `bson:"segment" json:"segment"`
	UserID    string           `bson:"user_id" json:"user_id"`
	Data      map[string]int64 `bson:"data" json:"data"`
}

func InsertCompaction(db *MongoDB, userID, dir, segment string, vals map[string]int64) error {
	if len(vals) == 0 {
		return errors.New("nil data")
	}

	return db.Run("compaction", func(c *mgo.Collection) error {
		return c.Insert(&Compaction{
			ID:        bson.NewObjectId(),
			UserID:    userID,
			Direction: dir,
			Segment:   segment,
			Data:      vals,
		})
	})
}

func GetCompaction(db *MongoDB, userID, dir, segment string) (map[string]int64, error) {
	query := bson.M{
		"user_id":   userID,
		"direction": dir,
		"segment":   segment,
	}
	res := &Compaction{}
	return res.Data, db.Run("compaction", func(c *mgo.Collection) error {
		err := c.Find(query).One(res)
		return err
	})
}

func DeleteCompaction(db *MongoDB, userID, dir, segment string) error {
	query := bson.M{
		"user_id":   userID,
		"direction": dir,
		"segment":   segment,
	}
	return db.Run("compaction", func(c *mgo.Collection) error {
		return c.Remove(query)
	})
}
