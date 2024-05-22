package store

import (
	"context"

	"github.com/dkinzler/linkboards/internal/auth"
	"github.com/dkinzler/linkboards/internal/boards/domain"

	"github.com/dkinzler/kit/errors"
)

// Default implementation of auth.AuthorizationStore that uses a datastore from the boards package.
// As the application is implemented right now, the roles a user has for a board are stored in a boards/domain.BoardUser instance that is persisted using a BoardDataStore.
type DefaultAuthorizationStore struct {
	ds domain.BoardDataStore
}

func NewDefaultAuthorizationStore(ds domain.BoardDataStore) auth.AuthorizationStore {
	return &DefaultAuthorizationStore{ds: ds}
}

func (d *DefaultAuthorizationStore) Roles(ctx context.Context, boardId string, userId string) ([]string, error) {
	user, err := d.ds.User(ctx, boardId, userId)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return nil, errors.New(err, "DefaultAuthorizationStore", errors.PermissionDenied)
		}
		return nil, errors.New(err, "DefaultAuthorizationStore", errors.Internal)
	}

	return []string{user.Role}, nil
}
