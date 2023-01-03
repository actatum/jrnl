package jrnl

import (
	"os"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"
)

const _testPassword = "password"

func TestNewJournal(t *testing.T) {
	f, closeFunc := mustNewTestFile(t)
	t.Cleanup(closeFunc)

	j, err := NewJournal(f.Name())
	if err != nil {
		t.Error(err)
	}

	err = j.Close()
	if err != nil {
		t.Error(err)
	}
}

func TestJournal_CreateEntry(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Entry
		wantErr bool
	}{
		{
			name:    "success",
			content: "a new journal entry for a new day",
			want: Entry{
				ID:         1,
				Content:    "a new journal entry for a new day",
				CreateTime: time.Now().UTC(),
				UpdateTime: time.Now().UTC(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, closeFunc := mustNewTestFile(t)
			t.Cleanup(closeFunc)

			j, jCloseFunc := mustNewAuthenticatedJournal(t, f.Name())
			t.Cleanup(jCloseFunc)

			got, err := j.CreateEntry(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEntry() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want, cmpopts.EquateApproxTime(5*time.Second)); diff != "" {
				t.Errorf("CreateEntry() (-got, +want):\n%s", diff)
			}
		})
	}
}

func TestJournal_EditEntry(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Entry
		wantErr bool
	}{
		{
			name:    "success",
			content: "i've been edited",
			want: Entry{
				ID:         1,
				Content:    "i've been edited",
				CreateTime: time.Now().UTC(),
				UpdateTime: time.Now().UTC(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, closeFunc := mustNewTestFile(t)
			t.Cleanup(closeFunc)

			j, jCloseFunc := mustNewAuthenticatedJournal(t, f.Name())
			t.Cleanup(jCloseFunc)

			e := mustCreateEntry(t, j, "i've been created")

			got, err := j.EditEntry(e.ID, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("EditEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want, cmpopts.EquateApproxTime(5*time.Second)); diff != "" {
				t.Errorf("EditEntry() (-got, +want):\n%s", diff)
			}
		})
	}
}

func TestJournal_ListEntries(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		seedData []string
		want     []Entry
		wantErr  bool
	}{
		{
			name:    "success",
			content: "i've been edited",
			seedData: []string{
				"first entry wow",
				"some stuff happened",
				"go is great",
			},
			want: []Entry{
				{
					ID:         3,
					Content:    "go is great",
					CreateTime: time.Now().UTC(),
					UpdateTime: time.Now().UTC(),
				},
				{
					ID:         2,
					Content:    "some stuff happened",
					CreateTime: time.Now().UTC(),
					UpdateTime: time.Now().UTC(),
				},
				{
					ID:         1,
					Content:    "first entry wow",
					CreateTime: time.Now().UTC(),
					UpdateTime: time.Now().UTC(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, closeFunc := mustNewTestFile(t)
			t.Cleanup(closeFunc)

			j, jCloseFunc := mustNewAuthenticatedJournal(t, f.Name())
			t.Cleanup(jCloseFunc)

			for _, content := range tt.seedData {
				mustCreateEntry(t, j, content)
			}

			got, err := j.ListEntries()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want, cmpopts.EquateApproxTime(5*time.Second)); diff != "" {
				t.Errorf("ListEntries() (-got, +want):\n%s", diff)
			}
		})
	}
}

func TestJournal_DeleteEntry(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, closeFunc := mustNewTestFile(t)
			t.Cleanup(closeFunc)

			j, jCloseFunc := mustNewAuthenticatedJournal(t, f.Name())
			t.Cleanup(jCloseFunc)

			e := mustCreateEntry(t, j, "i've been created")

			err := j.DeleteEntry(e.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = j.db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(journalBucketName))
				data := b.Get(itob(tt.id))
				if data != nil {
					t.Errorf("DeleteEntry() expected nil slice got %v", data)
				}

				return nil
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestJournal_CreatePassword(t *testing.T) {
	f, closeFunc := mustNewTestFile(t)
	t.Cleanup(closeFunc)

	j, err := NewJournal(f.Name())
	if err != nil {
		t.Error(err)
	}

	err = j.CreatePassword(_testPassword)
	if err != nil {
		t.Error(err)
	}

	err = j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(passwordBucketName))
		data := b.Get([]byte(passwordKey))

		if data == nil {
			t.Errorf("expected nil slice got %v", data)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestJournal_Auth(t *testing.T) {
	t.Run("password has been created", func(t *testing.T) {
		t.Run("correct password", func(t *testing.T) {
			f, closeFunc := mustNewTestFile(t)
			t.Cleanup(closeFunc)

			j, err := NewJournal(f.Name())
			if err != nil {
				t.Error(err)
			}

			err = j.CreatePassword(_testPassword)
			if err != nil {
				t.Error(err)
			}

			err = j.Auth(_testPassword)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("incorrect password", func(t *testing.T) {
			f, closeFunc := mustNewTestFile(t)
			t.Cleanup(closeFunc)

			j, err := NewJournal(f.Name())
			if err != nil {
				t.Error(err)
			}

			err = j.CreatePassword(_testPassword)
			if err != nil {
				t.Error(err)
			}

			err = j.Auth("some incorrect password")
			if err == nil {
				t.Errorf("expected error when inputting incorrect password")
			}
		})
	})

	t.Run("password hasn't been created", func(t *testing.T) {
		f, closeFunc := mustNewTestFile(t)
		t.Cleanup(closeFunc)

		j, err := NewJournal(f.Name())
		if err != nil {
			t.Error(err)
		}

		err = j.Auth(_testPassword)
		if err == nil {
			t.Errorf("expected error when authenticated before password has been creating")
		}
	})
}

func TestJournal_IsInitialized(t *testing.T) {
	t.Run("password has been created", func(t *testing.T) {
		f, closeFunc := mustNewTestFile(t)
		t.Cleanup(closeFunc)

		j, err := NewJournal(f.Name())
		if err != nil {
			t.Error(err)
		}

		err = j.CreatePassword(_testPassword)
		if err != nil {
			t.Error(err)
		}

		initialized, err := j.IsInitialized()
		if err != nil {
			t.Error(err)
		}

		if !initialized {
			t.Errorf("password has been created so initialized should be true")
		}
	})

	t.Run("password hasn't been created", func(t *testing.T) {
		f, closeFunc := mustNewTestFile(t)
		t.Cleanup(closeFunc)

		j, err := NewJournal(f.Name())
		if err != nil {
			t.Error(err)
		}

		initialized, err := j.IsInitialized()
		if err != nil {
			t.Error(err)
		}

		if initialized {
			t.Errorf("password hasn't been created so initialized should be false")
		}
	})
}

func mustCreateEntry(tb testing.TB, j *Journal, content string) Entry {
	tb.Helper()

	e, err := j.CreateEntry(content)
	if err != nil {
		tb.Fatal(err)
	}

	return e
}

func mustNewAuthenticatedJournal(tb testing.TB, dbPath string) (*Journal, func()) {
	tb.Helper()

	j, err := NewJournal(dbPath)
	if err != nil {
		tb.Fatal(err)
	}
	if err = j.CreatePassword(_testPassword); err != nil {
		tb.Fatal(err)
	}
	if err = j.Auth(_testPassword); err != nil {
		tb.Fatal(err)
	}

	return j, func() {
		if err := j.Close(); err != nil {
			tb.Fatal(err)
		}
	}
}

func mustNewTestFile(tb testing.TB) (*os.File, func()) {
	tb.Helper()

	f, err := os.CreateTemp("", "")
	if err != nil {
		tb.Fatal(err)
	}

	return f, func() {
		if err := f.Close(); err != nil {
			tb.Fatal(err)
		}
	}
}
