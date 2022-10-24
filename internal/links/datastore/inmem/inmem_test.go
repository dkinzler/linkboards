package inmem

import (
	"testing"

	"github.com/d39b/linkboards/internal/links/datastore"
)

func TestInmemLinkDataStore(t *testing.T) {
	s := NewInmemLinkDataStore()

	datastore.DatastoreTest(s, t)
}
