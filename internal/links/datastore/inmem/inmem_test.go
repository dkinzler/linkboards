package inmem

import (
	"go-sample/internal/links/datastore"
	"testing"
)

func TestInmemLinkDataStore(t *testing.T) {
	s := NewInmemLinkDataStore()

	datastore.DatastoreTest(s, t)
}
