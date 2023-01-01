package jrnl

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
)

const bucketName = "journal"

// Entry ...
type Entry struct {
	ID         int
	Content    string
	CreateTime time.Time
}

// Journal ...
type Journal struct {
	db *bolt.DB
}

// NewJournal returns a new instance of [Journal].
func NewJournal(dbPath string) (*Journal, error) {
	db, err := bolt.Open(dbPath, 0666, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Journal{
		db: db,
	}, nil
}

// Close closes the underlying boltdb.
func (j *Journal) Close() error {
	return j.db.Close()
}

// CreateEntry stores a new entry in the journal.
func (j *Journal) CreateEntry(content string) (Entry, error) {
	e := Entry{
		Content:    content,
		CreateTime: time.Now(),
	}

	err := j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		e.ID = int(id)

		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		encrypted, err := encrypt(buf, hashPassword(""))
		if err != nil {
			return err
		}

		return b.Put(itob(e.ID), encrypted)
	})
	if err != nil {
		return Entry{}, nil
	}

	return e, nil
}

// ListEntries lists all entries in the journal.
func (j *Journal) ListEntries() ([]Entry, error) {
	entries := make([]Entry, 0)

	err := j.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucketName))

		err := b.ForEach(func(k, v []byte) error {
			fmt.Printf("key=%s, value=%s\n", k, v)

			decrypted, err := decrypt(v, hashPassword(""))
			if err != nil {
				return err
			}

			var e Entry
			if err = json.Unmarshal(decrypted, &e); err != nil {
				return err
			}

			entries = append(entries, e)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID > entries[j].ID
	})

	return entries, nil
}

// DeleteEntry removes an entry from the journal.
func (j *Journal) DeleteEntry(id int) error {
	return j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Delete(itob(id))
	})
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
