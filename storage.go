package logshipper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	bolt "go.etcd.io/bbolt"
)

var db *bolt.DB
var bucketNames = []string{"events", "uploadedEvents"}
var dbPath = "./logs.db"

// InitStorage initializes the log database
func InitStorage() {
	var err error
	db, err = bolt.Open(dbPath, 0666, nil)
	if err != nil {
		log.Println("Fatal DB error")
		log.Fatalln(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		var err error

		for _, name := range bucketNames {
			_, err = tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Println("Fatal DB error")
		log.Fatalln(err)
	}

	fmt.Println("Storage initialized")

	go func() {
		fmt.Println("Recording events")
		for e := range Events {
			fmt.Println("Recording event: " + e.GetDetail())
			RecordEvent(e)
		}
	}()
}

// RecordEvent saves a log event
func RecordEvent(e Event) {
	db.Update(func(tx *bolt.Tx) error {
		events := tx.Bucket([]byte("events"))
		bz := e.GetBytes()
		id, _ := events.NextSequence()

		return events.Put(itob(id), bz)
	})
}

// GetEventsToUpload gets a batch of events that have not been uploaded - does not unmarshal
func GetEventsToUpload(limit int) ([][]byte, [][]byte) {
	eventSlice := [][]byte{}
	keySlice := [][]byte{}

	db.View(func(tx *bolt.Tx) error {
		events := tx.Bucket([]byte("events"))
		uploadedEvents := tx.Bucket([]byte("uploadedEvents"))
		uec := uploadedEvents.Cursor()
		ec := events.Cursor()

		lastUploadedKey, _ := uec.Last()

		if lastUploadedKey == nil {
			lastUploadedKey, _ = ec.First()
		}

		for k, v := ec.Seek(lastUploadedKey); k != nil && len(eventSlice) < limit; k, v = ec.Next() {
			eventSlice = append(eventSlice, v)
			keySlice = append(keySlice, k)
		}

		return nil
	})

	return eventSlice, keySlice
}

// MarkEventsUploaded marks an event as uploaded (along with the time), so it can be deleted later
func MarkEventsUploaded(keys [][]byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		now := uint64(time.Now().Unix())
		uploadedEvents := tx.Bucket([]byte("uploadedEvents"))
		var err error

		for _, key := range keys {
			err = uploadedEvents.Put(key, itob(now))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// LastUploadTime gets the `time.Time` of the last upload
func LastUploadTime() time.Time {
	var timestamp time.Time

	db.Update(func(tx *bolt.Tx) error {
		uploadedEvents := tx.Bucket([]byte("uploadedEvents"))
		uec := uploadedEvents.Cursor()
		lastUploadedKey, v := uec.Last()
		if lastUploadedKey == nil {
			return nil
		}
		timestamp = time.Unix(int64(btoi(v)), 0)
		return nil
	})

	return timestamp
}

// DeleteUploadedEvents fetches a batch of uploaded events, and deletes them
func DeleteUploadedEvents() error {
	return db.Update(func(tx *bolt.Tx) error {
		uploadedEvents := tx.Bucket([]byte("uploadedEvents"))
		events := tx.Bucket([]byte("events"))
		c := uploadedEvents.Cursor()

		var err error

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			err = events.Delete(k)
			if err != nil {
				return err
			}
			err = uploadedEvents.Delete(k)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// itob returns an 8-byte big endian encoding of v
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// btoi parses a uint64 from its 8-byte big endian encoding
func btoi(b []byte) uint64 {
	var v uint64
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &v)
	if err != nil {
		log.Fatalln(err)
	}
	return v
}
