package inmem

import (
	"linkboards/internal/boards/datastore"
	"testing"
)

func TestInmemBoardDataStore(t *testing.T) {
	s := NewInmemBoardDataStore()

	datastore.DatastoreTest(s, t)
}
