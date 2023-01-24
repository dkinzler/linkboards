package inmem

import (
	"testing"

	"github.com/dkinzler/linkboards/internal/boards/datastore"
)

func TestInmemBoardDataStore(t *testing.T) {
	s := NewInmemBoardDataStore()

	datastore.DatastoreTest(s, t)
}
