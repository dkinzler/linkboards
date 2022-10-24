package inmem

import (
	"testing"

	"github.com/d39b/linkboards/internal/boards/datastore"
)

func TestInmemBoardDataStore(t *testing.T) {
	s := NewInmemBoardDataStore()

	datastore.DatastoreTest(s, t)
}
