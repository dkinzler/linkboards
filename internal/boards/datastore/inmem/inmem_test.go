package inmem

import (
	"go-sample/internal/boards/datastore"
	"testing"
)

func TestInmemBoardDataStore(t *testing.T) {
	s := NewInmemBoardDataStore()

	datastore.DatastoreTest(s, t)
}
