package firestore

import (
	"testing"

	"github.com/dkinzler/linkboards/internal/boards/domain"

	"github.com/stretchr/testify/assert"
)

func TestFirestoreTypes(t *testing.T) {
	a := assert.New(t)

	user1 := domain.User{
		UserId: "u-1",
		Name:   "User 1",
	}
	user2 := domain.User{
		UserId: "u-2",
		Name:   "User 2",
	}

	a.Equal(user1, newDomainUser(newFsUser(user1)))
	a.Equal(user2, newDomainUser(newFsUser(user2)))

	board := domain.Board{
		BoardId:      "b-123",
		Name:         "name",
		Description:  "description",
		CreatedTime:  1234,
		CreatedBy:    user1,
		ModifiedTime: 5678,
		ModifiedBy:   user1,
	}

	a.Equal(board, newDomainBoard(newFsBoard(board)))

	boardUser1 := domain.BoardUser{
		User:         user1,
		Role:         "role1",
		CreatedTime:  42,
		InvitedBy:    user2,
		ModifiedTime: 42,
		ModifiedBy:   user2,
	}
	boardUser2 := domain.BoardUser{
		User:         user2,
		Role:         "role2",
		CreatedTime:  42,
		InvitedBy:    user1,
		ModifiedTime: 42,
		ModifiedBy:   user1,
	}

	a.Equal(boardUser1, newDomainBoardUser(newFsBoardUser(boardUser1, "b-123")))
	a.Equal(boardUser2, newDomainBoardUser(newFsBoardUser(boardUser2, "b-123")))

	boardInvite1 := domain.BoardInvite{
		InviteId:    "i-1",
		Role:        "role3",
		User:        user1,
		CreatedTime: 43,
		CreatedBy:   user2,
		ExpiresTime: 44,
	}
	boardInvite2 := domain.BoardInvite{
		InviteId:    "i-2",
		Role:        "role4",
		User:        user2,
		CreatedTime: 43,
		CreatedBy:   user1,
		ExpiresTime: 44,
	}

	a.Equal(boardInvite1, newDomainBoardInvite(newFsBoardInvite(boardInvite1, "b-123")))
	a.Equal(boardInvite2, newDomainBoardInvite(newFsBoardInvite(boardInvite2, "b-123")))

	boardWithUAndI := domain.BoardWithUsersAndInvites{
		Board:   board,
		Users:   []domain.BoardUser{boardUser1, boardUser2},
		Invites: []domain.BoardInvite{boardInvite1, boardInvite2},
	}

	fsBoardWithUAndI := fsBoardWithUsersAndInvites{
		Board: newFsBoard(board),
		Users: map[string]fsBoardUser{
			boardUser1.User.UserId: newFsBoardUser(boardUser1, "b-123"),
			boardUser2.User.UserId: newFsBoardUser(boardUser2, "b-123"),
		},
		Invites: map[string]fsBoardInvite{
			boardInvite1.InviteId: newFsBoardInvite(boardInvite1, "b-123"),
			boardInvite2.InviteId: newFsBoardInvite(boardInvite2, "b-123"),
		},
	}

	r := newDomainBoardWithUsersAndInvites(fsBoardWithUAndI)
	a.Equal(boardWithUAndI.Board, r.Board)
	a.ElementsMatch(boardWithUAndI.Users, r.Users)
	a.ElementsMatch(boardWithUAndI.Invites, r.Invites)
}
