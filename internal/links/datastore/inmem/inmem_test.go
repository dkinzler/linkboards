package inmem

import (
	"linkboards/internal/links/datastore"
	"testing"
)

func TestInmemLinkDataStore(t *testing.T) {
	s := NewInmemLinkDataStore()

	datastore.DatastoreTest(s, t)
}
