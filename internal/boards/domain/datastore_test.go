package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatastoreBoardUpdate(t *testing.T) {
	a := assert.New(t)

	board := Board{
		BoardId: "b-123",
		Name:    "Test Name",
	}
	bu1 := BoardUser{
		User: User{
			UserId: "u-1",
		},
	}
	bu2 := BoardUser{
		User: User{
			UserId: "u-2",
		},
	}
	bi1 := BoardInvite{
		InviteId: "i-1",
	}
	bi2 := BoardInvite{
		InviteId: "i-2",
	}

	u := NewDatastoreBoardUpdate("te")
	a.True(u.IsEmpty())

	u.WithBoard(board)
	a.False(u.IsEmpty())

	u.UpdateUser(bu1)
	u.UpdateUser(bu2)
	u.RemoveUser(bu1.User.UserId)
	u.RemoveUser(bu2.User.UserId)
	u.UpdateInvite(bi1)
	u.UpdateInvite(bi2)
	u.RemoveInvite(bi1.InviteId)
	u.RemoveInvite(bi2.InviteId)

	a.False(u.IsEmpty())
	a.Equal("te", u.TransactionExpecation)
	a.Equal(true, u.UpdateBoard)
	a.Equal(board, u.Board)
	a.ElementsMatch([]BoardUser{bu1, bu2}, u.UpdateUsers)
	a.ElementsMatch([]string{"u-1", "u-2"}, u.RemoveUsers)
	a.ElementsMatch([]BoardInvite{bi1, bi2}, u.UpdateInvites)
	a.ElementsMatch([]string{"i-1", "i-2"}, u.RemoveInvites)
}

func TestDatastoreQueryParams(t *testing.T) {
	a := assert.New(t)

	qp := NewQueryParams()
	a.Equal(20, qp.Limit)

	// value too large
	qp = qp.WithLimit(1000)
	a.Equal(20, qp.Limit)

	qp = qp.WithLimit(50)
	a.Equal(50, qp.Limit)

	qp = qp.WithCursor(133742)
	a.EqualValues(133742, qp.Cursor)
}
