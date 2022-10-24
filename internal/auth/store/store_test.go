package store

import (
	"context"
	"testing"

	"github.com/d39b/linkboards/internal/auth"
	"github.com/d39b/linkboards/internal/boards/datastore/inmem"
	"github.com/d39b/linkboards/internal/boards/domain"

	"github.com/d39b/kit/errors"
	"github.com/stretchr/testify/assert"
)

func TestDefaultAuthorizationStore(t *testing.T) {
	a := assert.New(t)

	bds := inmem.NewInmemBoardDataStore()
	err := bds.UpdateBoard(context.Background(), "b-123", domain.NewDatastoreBoardUpdate(nil).WithBoard(domain.Board{
		BoardId: "b-123",
	}).UpdateUser(domain.BoardUser{
		User: domain.User{
			UserId: "u-1",
		},
		Role: auth.BoardRoleOwner,
	}).UpdateUser(domain.BoardUser{
		User: domain.User{
			UserId: "u-2",
		},
		Role: auth.BoardRoleViewer,
	}))
	a.Nil(err)

	store := NewDefaultAuthorizationStore(bds)
	ctx := context.Background()

	// returns PermissionDenied when board not found or user not on board
	roles, err := store.Roles(ctx, "b-456", "u-1")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))
	a.Empty(roles)

	roles, err = store.Roles(ctx, "b-123", "u-3")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))
	a.Empty(roles)

	// should return correct roles
	roles, err = store.Roles(ctx, "b-123", "u-1")
	a.Nil(err)
	a.Len(roles, 1)
	a.Contains(roles, auth.BoardRoleOwner)

	roles, err = store.Roles(ctx, "b-123", "u-2")
	a.Nil(err)
	a.Len(roles, 1)
	a.Contains(roles, auth.BoardRoleViewer)

}
