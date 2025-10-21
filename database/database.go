package database

import (
	"time"

	"github.com/asdine/storm/v3"
	"github.com/hashicorp/go-hclog"
	bolt "go.etcd.io/bbolt"
)

// StartDB instantiates the database
func StartDB(dbdir string) (database *storm.DB) {
	dbpath := dbdir + "/nightlight-cloud.db"
	db, err := storm.Open(dbpath, storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}))
	if err != nil {
		hclog.Default().Named("database").Error(err.Error())
	}
	return db
}

// DeleteDBRecord deletes a single database record
func DeleteDBRecord(db *bolt.DB, bucket string, key string) {

	if err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Delete([]byte(key))
	}); err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
}

// GetDBRecord gets a single database record
func GetDBRecord(db *bolt.DB, bucket string, key string) (data string) {
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		v := b.Get([]byte(key))
		data = string(v)
		return nil
	})
	if err != nil {
		hclog.Default().Named("core").Error(err.Error())
	}
	return data
}
