package datastore

import (
	"context"
	"go-sample/internal/boards/domain"
	"testing"
	"time"

	"github.com/d39b/kit/errors"

	"github.com/stretchr/testify/assert"
)

// General tests that can be run against any implementation of BoardDataStore.
func DatastoreTest(ds domain.BoardDataStore, t *testing.T) {
	GeneralTests(ds, t)
	QueryCursorTest(ds, t)
}

func GeneralTests(ds domain.BoardDataStore, t *testing.T) {
	a := assert.New(t)

	board := domain.Board{
		BoardId: "b-123",
	}
	boardUser1 := domain.BoardUser{
		User: domain.User{UserId: "u-1"},
	}
	boardUser2 := domain.BoardUser{
		User: domain.User{UserId: "u-2"},
	}

	boardInvite1 := domain.BoardInvite{
		InviteId: "i-1",
		User:     domain.User{UserId: "u-3"},
	}
	boardInvite2 := domain.BoardInvite{
		InviteId: "i-2",
		User:     domain.User{UserId: "u-4"},
	}
	ctx, cancel := getContext()
	defer cancel()
	b, _, err := ds.Board(ctx, board.BoardId)
	a.Empty(b)
	a.NotNil(err)
	a.True(errors.IsNotFoundError(err))

	err = ds.UpdateBoard(ctx, board.BoardId, domain.NewDatastoreBoardUpdate(nil).
		WithBoard(board).
		UpdateUser(boardUser1).
		UpdateUser(boardUser2).
		UpdateInvite(boardInvite1).
		UpdateInvite(boardInvite2))
	a.Nil(err)

	b, boardTe, err := ds.Board(ctx, board.BoardId)
	a.Equal(board, b.Board)
	a.Nil(err)
	a.NotNil(boardTe)
	a.Equal(2, b.InviteCount())
	a.Equal(2, b.UserCount())

	// check that transaction expectations work
	// we will update the users of the board
	// and if we try to perform another update afterwards, with the same old transaction expectations
	// it shouldn't work

	err = ds.UpdateBoard(ctx, board.BoardId, domain.NewDatastoreBoardUpdate(boardTe).WithBoard(domain.Board{BoardId: board.BoardId, Name: "test1"}))
	a.Nil(err)

	// now trying to update the board with the old transaction expectations should fail
	err = ds.UpdateBoard(ctx, board.BoardId, domain.NewDatastoreBoardUpdate(boardTe).WithBoard(domain.Board{BoardId: board.BoardId, Name: "test2"}))
	a.NotNil(err)
	a.True(errors.IsFailedPreconditionError(err))

	// if we get new expecations it should work
	_, te, err := ds.Board(ctx, board.BoardId)
	a.Nil(err)
	a.NotNil(te)
	err = ds.UpdateBoard(ctx, board.BoardId, domain.NewDatastoreBoardUpdate(te).WithBoard(domain.Board{BoardId: board.BoardId, Name: "test2"}))
	a.Nil(err)

	err = ds.UpdateBoard(ctx, "b-456", domain.NewDatastoreBoardUpdate(nil).WithBoard(domain.Board{BoardId: "b-456"}).UpdateUser(domain.BoardUser{User: domain.User{UserId: "u-2"}}))
	a.Nil(err)
	err = ds.UpdateBoard(ctx, "b-789", domain.NewDatastoreBoardUpdate(nil).WithBoard(domain.Board{BoardId: "b-789"}).UpdateUser(domain.BoardUser{User: domain.User{UserId: "u-2"}}))
	a.Nil(err)
	err = ds.UpdateBoard(ctx, "b-012", domain.NewDatastoreBoardUpdate(nil).WithBoard(domain.Board{BoardId: "b-012"}).UpdateUser(domain.BoardUser{User: domain.User{UserId: "u-2"}}))
	a.Nil(err)

	boardUser, err := ds.User(ctx, "b-789", "u-4")
	a.Empty(boardUser)
	a.NotNil(err)
	a.True(errors.IsNotFoundError(err))

	boardUser, err = ds.User(ctx, "b-789", "u-2")
	a.Nil(err)
	a.Equal("u-2", boardUser.User.UserId)

	boards, err := ds.BoardsForUser(ctx, "u-1", domain.NewQueryParams())
	a.Nil(err)
	a.Len(boards, 1)
	a.Equal("b-123", boards[0].BoardId)

	boards, err = ds.BoardsForUser(ctx, "u-2", domain.NewQueryParams())
	a.Nil(err)
	a.Len(boards, 4)
	a.ElementsMatch([]string{"b-123", "b-456", "b-789", "b-012"}, []string{boards[0].BoardId, boards[1].BoardId, boards[2].BoardId, boards[3].BoardId})

	err = ds.UpdateBoard(ctx, "b-456", domain.NewDatastoreBoardUpdate(nil).UpdateInvite(domain.BoardInvite{InviteId: "iii", User: domain.User{UserId: "u-4"}}))
	a.Nil(err)
	err = ds.UpdateBoard(ctx, "b-789", domain.NewDatastoreBoardUpdate(nil).UpdateInvite(domain.BoardInvite{InviteId: "iiii", User: domain.User{UserId: "u-4"}}))
	a.Nil(err)

	invites, err := ds.InvitesForUser(ctx, "u-1", domain.NewQueryParams())
	a.Nil(err)
	a.Len(invites, 0)

	invites, err = ds.InvitesForUser(ctx, "u-4", domain.NewQueryParams())
	a.Nil(err)
	a.Len(invites, 3)
	a.Contains(invites, "b-123")
	a.Contains(invites, "b-456")
	a.Contains(invites, "b-789")

	err = ds.DeleteBoard(ctx, "b-123")
	a.Nil(err)
	err = ds.DeleteBoard(ctx, "b-789")
	a.Nil(err)
	err = ds.DeleteBoard(ctx, "b-456")
	a.Nil(err)
	err = ds.DeleteBoard(ctx, "b-012")
	a.Nil(err)

	boards, err = ds.BoardsForUser(ctx, "u-1", domain.NewQueryParams())
	a.Nil(err)
	a.Len(boards, 0)

	boards, err = ds.BoardsForUser(ctx, "u-2", domain.NewQueryParams())
	a.Nil(err)
	a.Len(boards, 0)

	invites, err = ds.InvitesForUser(ctx, "u-4", domain.NewQueryParams())
	a.Nil(err)
	a.Len(invites, 0)
}

func QueryCursorTest(ds domain.BoardDataStore, t *testing.T) {
	a := assert.New(t)

	board1 := domain.Board{
		BoardId: "b-123",
	}
	boardUser1 := domain.BoardUser{
		User: domain.User{
			UserId: "u-1",
		},
		CreatedTime: 1000,
	}
	boardInvite1 := domain.BoardInvite{
		InviteId:    "i-1",
		User:        domain.User{UserId: "u-2"},
		CreatedTime: 1000,
	}
	board2 := domain.Board{
		BoardId: "b-456",
	}
	boardUser2 := domain.BoardUser{
		User: domain.User{
			UserId: "u-1",
		},
		CreatedTime: 2000,
	}
	boardInvite2 := domain.BoardInvite{
		InviteId:    "i-2",
		User:        domain.User{UserId: "u-2"},
		CreatedTime: 2000,
	}
	board3 := domain.Board{
		BoardId: "b-789",
	}
	boardUser3 := domain.BoardUser{
		User: domain.User{
			UserId: "u-1",
		},
		CreatedTime: 3000,
	}
	boardInvite3 := domain.BoardInvite{
		InviteId:    "i-3",
		User:        domain.User{UserId: "u-2"},
		CreatedTime: 3000,
	}

	ctx, cancel := getContext()
	defer cancel()

	err := ds.UpdateBoard(ctx, board1.BoardId, domain.NewDatastoreBoardUpdate(nil).
		WithBoard(board1).
		UpdateUser(boardUser1).
		UpdateInvite(boardInvite1))
	a.Nil(err)
	err = ds.UpdateBoard(ctx, board2.BoardId, domain.NewDatastoreBoardUpdate(nil).
		WithBoard(board2).
		UpdateUser(boardUser2).
		UpdateInvite(boardInvite2))
	a.Nil(err)
	err = ds.UpdateBoard(ctx, board3.BoardId, domain.NewDatastoreBoardUpdate(nil).
		WithBoard(board3).
		UpdateUser(boardUser3).
		UpdateInvite(boardInvite3))
	a.Nil(err)

	boards, err := ds.BoardsForUser(ctx, "u-1", domain.NewQueryParams().WithLimit(2))
	a.Nil(err)
	// should return newest 2 boards, i.e. board3 and board2
	a.Len(boards, 2)
	a.Equal(board3.BoardId, boards[0].BoardId)
	a.Equal(board2.BoardId, boards[1].BoardId)

	boards, err = ds.BoardsForUser(ctx, "u-1", domain.NewQueryParams().WithCursor(2500))
	a.Nil(err)
	// this should return board 2 and 1
	a.Len(boards, 2)
	a.Equal(board2.BoardId, boards[0].BoardId)
	a.Equal(board1.BoardId, boards[1].BoardId)

	boards, err = ds.BoardsForUser(ctx, "u-1", domain.NewQueryParams().WithCursor(boardUser1.CreatedTime))
	a.Nil(err)
	// this should only return board 1
	a.Len(boards, 1)
	a.Equal(board1.BoardId, boards[0].BoardId)

	invites, err := ds.InvitesForUser(ctx, "u-2", domain.NewQueryParams().WithLimit(2))
	a.Nil(err)
	// should return newest 2 invites
	a.Len(invites, 2)
	a.Contains(invites, board3.BoardId)
	a.Contains(invites, board2.BoardId)

	invites, err = ds.InvitesForUser(ctx, "u-2", domain.NewQueryParams().WithCursor(2500))
	a.Nil(err)
	// this should return invites 2 and 1
	a.Len(invites, 2)
	a.Contains(invites, board2.BoardId)
	a.Contains(invites, board1.BoardId)

	invites, err = ds.InvitesForUser(ctx, "u-2", domain.NewQueryParams().WithCursor(boardInvite1.CreatedTime))
	a.Nil(err)
	// this should only return invite 1
	a.Len(invites, 1)
	a.Contains(invites, board1.BoardId)
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}
