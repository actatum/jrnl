package jrnl

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

const (
	journalBucketName  = "journal"
	passwordBucketName = "password"
	passwordKey        = "pw"
)

// Entry is an individual journal entry.
type Entry struct {
	ID         int
	Content    string
	CreateTime time.Time
	UpdateTime time.Time
}

// Journal manages persisting journal entries.
type Journal struct {
	db             *bolt.DB
	hashedPassword string
}

// NewJournal returns a new instance of Journal.
func NewJournal(dbPath string) (*Journal, error) {
	db, err := bolt.Open(dbPath, 0666, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists([]byte(journalBucketName)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err = tx.CreateBucketIfNotExists([]byte(passwordBucketName)); err != nil {
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
	now := time.Now()
	e := Entry{
		Content:    content,
		CreateTime: now,
		UpdateTime: now,
	}

	err := j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(journalBucketName))
		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		e.ID = int(id)

		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		encrypted, err := encrypt([]byte(j.hashedPassword), buf)
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

// EditEntry edits an existing entry
func (j *Journal) EditEntry(id int, content string) (Entry, error) {
	e := Entry{
		ID:         id,
		Content:    content,
		UpdateTime: time.Now(),
	}

	err := j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(journalBucketName))

		data := b.Get(itob(id))
		decrypted, err := decrypt([]byte(j.hashedPassword), data)
		if err != nil {
			return err
		}

		var currentEntry Entry
		if err = json.Unmarshal(decrypted, &currentEntry); err != nil {
			return err
		}

		e.CreateTime = currentEntry.CreateTime

		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		encrypted, err := encrypt([]byte(j.hashedPassword), buf)
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
		b := tx.Bucket([]byte(journalBucketName))

		err := b.ForEach(func(k, v []byte) error {
			decrypted, err := decrypt([]byte(j.hashedPassword), v)
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
		b := tx.Bucket([]byte(journalBucketName))
		return b.Delete(itob(id))
	})
}

// CreatePassword stores a user's password.
func (j *Journal) CreatePassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(passwordBucketName))
		return b.Put([]byte(passwordKey), hash)
	})
}

// Auth authenticates a user to their journal.
func (j *Journal) Auth(password string) error {
	return j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(passwordBucketName))
		hash := b.Get([]byte(passwordKey))

		if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
			switch {
			case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
				return fmt.Errorf("incorrect password")
			}
			return err
		}

		j.hashedPassword = hashPassword(password)

		return nil
	})
}

// IsInitialized tells us if the journal has been password protected.
func (j *Journal) IsInitialized() (bool, error) {
	initialized := false
	err := j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(passwordBucketName))
		data := b.Get([]byte(passwordKey))
		if data != nil {
			initialized = true
		}

		return nil
	})
	if err != nil {
		return initialized, err
	}

	return initialized, nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
