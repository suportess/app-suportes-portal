package store

import (
	"fmt"
	"log"
	"time"

	"github.com/asdine/storm/v3"
	bolt "go.etcd.io/bbolt"
)

// DB wraps a storm.DB instance for embedded BoltDB storage.
type DB struct {
	Storm *storm.DB
}

// Open opens (or creates) the BoltDB file at the given path.
func Open(path string, timeout time.Duration) (*DB, error) {
	boltOpts := &bolt.Options{Timeout: timeout}
	s, err := storm.Open(path, storm.BoltOptions(0600, boltOpts))
	if err != nil {
		return nil, fmt.Errorf("failed to open store at %q: %w", path, err)
	}
	log.Printf("[store] opened database at %s", path)
	return &DB{Storm: s}, nil
}

// Close closes the underlying BoltDB.
func (db *DB) Close() error {
	return db.Storm.Close()
}
