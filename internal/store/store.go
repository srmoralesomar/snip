package store

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketClips = []byte("clips")
	bucketMeta  = []byte("meta")
	keyLastHash = []byte("last_hash")
)

// Clip represents a stored clipboard entry.
type Clip struct {
	ID        uint64    `json:"id"`
	Content   string    `json:"content"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}

// Store manages clipboard history in a BoltDB database.
type Store struct {
	db *bolt.DB
}

// DefaultPath returns the default database path (~/.snip/history.db).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".snip", "history.db"), nil
}

// New opens or creates the store at the given path.
// It creates the parent directory if it does not exist.
func New(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketClips); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists(bucketMeta)
		return err
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Save stores a new clip. If the content is identical to the last stored entry,
// it is silently ignored (no consecutive duplicates).
func (s *Store) Save(content string) error {
	if content == "" {
		return nil
	}
	hash := hashContent(content)

	return s.db.Update(func(tx *bolt.Tx) error {
		meta := tx.Bucket(bucketMeta)
		if last := meta.Get(keyLastHash); string(last) == hash {
			return nil // duplicate — skip
		}

		clips := tx.Bucket(bucketClips)
		id, err := clips.NextSequence()
		if err != nil {
			return fmt.Errorf("next sequence: %w", err)
		}

		clip := Clip{
			ID:        id,
			Content:   content,
			Hash:      hash,
			Timestamp: time.Now(),
		}
		data, err := json.Marshal(clip)
		if err != nil {
			return fmt.Errorf("marshal clip: %w", err)
		}

		if err := clips.Put(encodeID(id), data); err != nil {
			return fmt.Errorf("put clip: %w", err)
		}
		return meta.Put(keyLastHash, []byte(hash))
	})
}

// List returns up to n clips, newest first. If n <= 0, all clips are returned.
func (s *Store) List(n int) ([]Clip, error) {
	var clips []Clip

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketClips)
		c := b.Cursor()

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var clip Clip
			if err := json.Unmarshal(v, &clip); err != nil {
				return fmt.Errorf("unmarshal clip: %w", err)
			}
			clips = append(clips, clip)
			if n > 0 && len(clips) >= n {
				break
			}
		}
		return nil
	})

	return clips, err
}

// Get returns the clip with the given ID, or an error if not found.
func (s *Store) Get(id uint64) (*Clip, error) {
	var clip Clip

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketClips)
		v := b.Get(encodeID(id))
		if v == nil {
			return fmt.Errorf("clip %d not found", id)
		}
		return json.Unmarshal(v, &clip)
	})
	if err != nil {
		return nil, err
	}
	return &clip, nil
}

// Delete removes the clip with the given ID.
func (s *Store) Delete(id uint64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketClips)
		if b.Get(encodeID(id)) == nil {
			return fmt.Errorf("clip %d not found", id)
		}
		return b.Delete(encodeID(id))
	})
}

// Prune removes the oldest clips so that at most maxCount clips remain.
// If maxCount <= 0, Prune is a no-op.
func (s *Store) Prune(maxCount int) error {
	if maxCount <= 0 {
		return nil
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketClips)
		total := b.Stats().KeyN
		excess := total - maxCount
		if excess <= 0 {
			return nil
		}

		c := b.Cursor()
		var toDelete [][]byte
		for k, _ := c.First(); k != nil && len(toDelete) < excess; k, _ = c.Next() {
			key := make([]byte, len(k))
			copy(key, k)
			toDelete = append(toDelete, key)
		}
		for _, k := range toDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	})
}

// Count returns the total number of stored clips.
func (s *Store) Count() (int, error) {
	var n int
	err := s.db.View(func(tx *bolt.Tx) error {
		n = tx.Bucket(bucketClips).Stats().KeyN
		return nil
	})
	return n, err
}

// encodeID encodes a uint64 as a big-endian 8-byte key so BoltDB cursor
// iteration is in insertion (ascending ID) order.
func encodeID(id uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)
	return b
}

// hashContent returns the hex-encoded SHA-256 hash of the content string.
func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum)
}
