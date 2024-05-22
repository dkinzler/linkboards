package application

import (
	"context"
	"testing"
	stdtime "time"

	"github.com/dkinzler/linkboards/internal/auth"
	"github.com/dkinzler/linkboards/internal/auth/store"
	"github.com/dkinzler/linkboards/internal/boards/datastore/inmem"
	"github.com/dkinzler/linkboards/internal/boards/domain"

	"github.com/dkinzler/kit/errors"
	"github.com/dkinzler/kit/time"

	"github.com/stretchr/testify/assert"
)

var testUser1 = auth.User{
	UserId: "user-1",
	Name:   "User One",
}

func newTestDatastores() (domain.BoardDataStore, auth.AuthorizationStore) {
	ds := inmem.NewInmemBoardDataStore()
	as := store.NewDefaultAuthorizationStore(ds)
	return ds, as
}

func TestUnauthenticatedUsersDenied(t *testing.T) {
	a := assert.New(t)

	// Unauthenticated users are denied

	ctx := context.Background()

	service := NewBoardApplicationService(newTestDatastores())

	_, err := service.CreateBoard(ctx, NewBoard{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	err = service.DeleteBoard(ctx, "abc")
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	err = service.DeleteInvite(ctx, "abc", "i")
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	err = service.RemoveUser(ctx, "abc", "user")
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	err = service.RespondToInvite(ctx, "abc", "i", InviteResponse{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.Board(ctx, "abc")
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.Boards(ctx, QueryParams{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.CreateInvite(ctx, "abc", NewInvite{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.EditBoard(ctx, "abc", BoardEdit{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.EditBoardUser(ctx, "abc", "user", BoardUserEdit{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.Invites(ctx, QueryParams{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))
}

func TestAuthorization(t *testing.T) {
	a := assert.New(t)

	ds := inmem.NewInmemBoardDataStore()
	err := ds.UpdateBoard(context.Background(), "b-123", domain.NewDatastoreBoardUpdate(nil).
		WithBoard(domain.Board{BoardId: "b-123"}).
		UpdateUser(domain.BoardUser{User: toDomainUser(testUser1), Role: "testRole"}))
	a.Nil(err)

	// This will create a new service that uses an AuthorizationChecker
	// that maps the "testRole" role to all scopes except the given one.
	newService := func(withoutScope auth.Scope) BoardApplicationService {
		as := allScopes()
		scopes := make([]auth.Scope, 0, len(as)-1)
		for _, scope := range as {
			if scope != withoutScope {
				scopes = append(scopes, scope)
			}
		}

		rts := map[string][]auth.Scope{
			"testRole": scopes,
		}

		authStore := store.NewDefaultAuthorizationStore(ds)
		service := &boardApplicationService{
			boardService:   domain.NewBoardService(ds, nil),
			boardDataStore: ds,
			authChecker:    auth.NewAuthorizationChecker(rts, authStore),
		}
		return service
	}

	ctx := auth.ContextWithUser(context.Background(), testUser1)

	err = newService(deleteBoardScope).DeleteBoard(ctx, "b-123")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	_, err = newService(editBoardScope).EditBoard(ctx, "b-123", BoardEdit{})
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	_, err = newService(viewBoardScope).Board(ctx, "b-123")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	_, err = newService(createInviteScope).CreateInvite(ctx, "b-123", NewInvite{})
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	err = newService(deleteInviteScope).DeleteInvite(ctx, "b-123", "i-1")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	err = newService(removeUserFromBoardScope).RemoveUser(ctx, "b-123", "u-1")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	_, err = newService(editBoardUserScope).EditBoardUser(ctx, "b-123", "u-1", BoardUserEdit{})
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))
}

func TestBoardReturnsUsersAndInvitesOnlyIfAuthorized(t *testing.T) {
	a := assert.New(t)

	board := domain.Board{BoardId: "b-123", CreatedBy: toDomainUser(testUser1), CreatedTime: 41}
	boardUser := domain.BoardUser{User: toDomainUser(testUser1), Role: "testRole"}
	boardInvite := domain.BoardInvite{Role: "testRole", CreatedTime: 42}

	ds := inmem.NewInmemBoardDataStore()
	err := ds.UpdateBoard(context.Background(), "b-123", domain.NewDatastoreBoardUpdate(nil).
		WithBoard(board).
		UpdateUser(boardUser).
		UpdateInvite(boardInvite))
	a.Nil(err)

	// Creats a new service with an AuthorizationChecker that assigns the given scopes
	// to the user.
	newService := func(scopes []auth.Scope) BoardApplicationService {
		rts := map[string][]auth.Scope{
			"testRole": scopes,
		}

		authStore := store.NewDefaultAuthorizationStore(ds)
		service := &boardApplicationService{
			boardService:   domain.NewBoardService(ds, nil),
			boardDataStore: ds,
			authChecker:    auth.NewAuthorizationChecker(rts, authStore),
		}
		return service
	}

	ctx := auth.ContextWithUser(context.Background(), testUser1)

	// response should only include the board, no users or invites
	b, err := newService([]auth.Scope{viewBoardScope}).Board(ctx, "b-123")
	a.Nil(err)
	a.Equal(board.BoardId, b.BoardId)
	a.Equal(board.CreatedBy, b.CreatedBy)
	a.Equal(board.CreatedTime, b.CreatedTime)
	a.Empty(b.Users)
	a.Empty(b.Invites)

	// response should include users but not invites
	b, err = newService([]auth.Scope{viewBoardScope, viewBoardUsersScope}).Board(ctx, "b-123")
	a.Nil(err)
	a.NotEmpty(b.Users)
	u := b.Users[0]
	a.Equal(toDomainUser(testUser1), u.User)
	a.Equal("testRole", u.Role)
	a.Empty(b.Invites)

	// response should include invites but not users
	b, err = newService([]auth.Scope{viewBoardScope, viewBoardInvitesScope}).Board(ctx, "b-123")
	a.Nil(err)
	a.NotEmpty(b.Invites)
	i := b.Invites[0]
	a.Equal("testRole", i.Role)
	a.Equal(int64(42), i.CreatedTime)
	a.Empty(b.Users)

	// response should include users and invites
	b, err = newService([]auth.Scope{viewBoardScope, viewBoardInvitesScope, viewBoardUsersScope}).Board(ctx, "b-123")
	a.Nil(err)
	a.Len(b.Invites, 1)
	a.Len(b.Users, 1)
}

func TestEditTypes(t *testing.T) {
	a := assert.New(t)

	a.True(BoardEdit{}.IsEmpty())

	name := "abc"
	description := "xyz"
	be := BoardEdit{
		Name:        &name,
		Description: &description,
	}
	a.False(be.IsEmpty())

	d := be.ToDomainBoardEdit()
	a.True(d.UpdateName)
	a.Equal("abc", d.Name)
	a.True(d.UpdateDescription)
	a.Equal("xyz", d.Description)

	a.True(BoardUserEdit{}.IsEmpty())
	role := "role123"
	bue := BoardUserEdit{
		Role: &role,
	}
	a.False(bue.IsEmpty())

	d2 := bue.ToDomainBoardUserEdit()
	a.True(d2.UpdateRole)
	a.Equal("role123", d2.Role)
}

var defaultTime = stdtime.Date(2022, 04, 04, 0, 0, 0, 0, stdtime.UTC)
var defaultTimeUnix = defaultTime.UnixNano()

func TestCreateBoard(t *testing.T) {
	a := assert.New(t)

	time.TimeFunc = func() stdtime.Time {
		return defaultTime
	}

	ctx := auth.ContextWithUser(context.Background(), auth.User{
		UserId: "u-123",
		Name:   "Testi Tester",
	})
	service := NewBoardApplicationService(newTestDatastores())

	board, err := service.CreateBoard(ctx, NewBoard{
		Name:        "Board name",
		Description: "Board description",
	})
	a.Nil(err)
	a.NotEmpty(board.BoardId)
	a.Equal("Board name", board.Name)
	a.Equal("Board description", board.Description)
	a.Equal(domain.User{
		UserId: "u-123",
		Name:   "Testi Tester",
	}, board.CreatedBy)
	a.Equal(domain.User{
		UserId: "u-123",
		Name:   "Testi Tester",
	}, board.ModifiedBy)
	a.Equal(defaultTimeUnix, board.CreatedTime)
	a.Equal(defaultTimeUnix, board.ModifiedTime)
	a.Len(board.Invites, 0)
	a.Len(board.Users, 1)
	creator := board.Users[0]
	a.Equal(domain.User{
		UserId: "u-123",
		Name:   "Testi Tester",
	}, creator.User)
	a.Equal(auth.BoardRoleOwner, creator.Role)
	a.Equal(defaultTimeUnix, creator.CreatedTime)
	a.Equal(defaultTimeUnix, creator.ModifiedTime)
}
