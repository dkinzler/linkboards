package domain

import (
	"strings"
	"testing"
	stdtime "time"

	"github.com/d39b/linkboards/internal/auth"

	"github.com/d39b/kit/errors"
	"github.com/d39b/kit/time"

	"github.com/stretchr/testify/assert"
)

var defaultTime = stdtime.Date(2022, 1, 1, 0, 0, 0, 0, stdtime.UTC)
var defaultTimeUnix = defaultTime.UnixNano()

func TestBoardCreation(t *testing.T) {
	a := assert.New(t)

	time.TimeFunc = func() stdtime.Time {
		return defaultTime
	}

	//empty name
	board, err := NewBoard("", "", User{})
	a.Empty(board)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errBoardNameEmpty))

	//name too long
	var sb strings.Builder
	for i := 0; i <= descriptionMaxLength; i++ {
		sb.WriteString("a")
	}

	longString := sb.String()

	board, err = NewBoard(longString, "", User{})
	a.Empty(board)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errBoardNameTooLong))

	board, err = NewBoard("name", longString, User{})
	a.Empty(board)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errBoardDescriptionTooLong))

	//this should work
	board, err = NewBoard("name", "description", User{UserId: "abc", Name: "xyz"})
	a.Nil(err)
	a.Nil(board.IsValid())
	a.True(len(board.BoardId) > 10)
	a.Equal("name", board.Name)
	a.Equal("description", board.Description)
	a.Equal(defaultTimeUnix, board.CreatedTime)
	a.Equal(User{UserId: "abc", Name: "xyz"}, board.CreatedBy)
	a.Equal(defaultTimeUnix, board.ModifiedTime)
	a.Equal(User{UserId: "abc", Name: "xyz"}, board.ModifiedBy)
}

func TestBoardUserCreation(t *testing.T) {
	a := assert.New(t)

	time.TimeFunc = func() stdtime.Time {
		return defaultTime
	}

	user, err := NewBoardUser(User{}, auth.BoardRoleOwner, User{})
	a.NotNil(err)
	a.Empty(user)
	a.True(errors.IsInvalidArgumentError(err))

	user, err = NewBoardUser(User{UserId: "uid", Name: "Testi Tester"}, "invalidrole", User{UserId: "abc", Name: "xyz"})
	a.NotNil(err)
	a.Empty(user)
	a.True(errors.HasPublicCode(err, errInvalidRole))

	user, err = NewBoardUser(User{UserId: "uid", Name: "Testi Tester"}, auth.BoardRoleOwner, User{UserId: "abc", Name: "xyz"})
	a.Nil(err)
	a.Equal(User{UserId: "uid", Name: "Testi Tester"}, user.User)
	a.Equal(auth.BoardRoleOwner, user.Role)
	a.Equal(User{UserId: "abc", Name: "xyz"}, user.InvitedBy)
	a.Equal(User{UserId: "abc", Name: "xyz"}, user.ModifiedBy)
	a.Equal(defaultTimeUnix, user.CreatedTime)
	a.Equal(defaultTimeUnix, user.ModifiedTime)

	err = user.ChangeRole("invalidrole", User{})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errInvalidRole))

	err = user.ChangeRole(auth.BoardRoleEditor, User{UserId: "123"})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errCannotChangeRoleOfOwner))

	user, err = NewBoardUser(User{UserId: "uid", Name: "Testi Tester"}, auth.BoardRoleViewer, User{UserId: "abc", Name: "xyz"})
	a.Nil(err)
	err = user.ChangeRole(auth.BoardRoleEditor, User{UserId: "123"})
	a.Nil(err)
	a.Equal(auth.BoardRoleEditor, user.Role)
	a.Equal(User{UserId: "123"}, user.ModifiedBy)
}

func TestInviteCreation(t *testing.T) {
	a := assert.New(t)

	time.TimeFunc = func() stdtime.Time {
		return defaultTime
	}

	invite, err := NewBoardInvite("invalidrole", User{}, User{}, 1000*stdtime.Second)
	a.NotNil(err)
	a.Empty(invite)
	a.True(errors.HasPublicCode(err, errInvalidRole))

	invite, err = NewBoardInvite(auth.BoardRoleEditor, User{UserId: "u-123"}, User{UserId: "u-456"}, 42*stdtime.Second)
	a.Nil(err)
	a.True(len(invite.InviteId) > 10)
	a.Equal(auth.BoardRoleEditor, invite.Role)
	a.Equal("u-456", invite.User.UserId)
	a.Equal("u-123", invite.CreatedBy.UserId)
	a.Equal(defaultTimeUnix, invite.CreatedTime)
	a.Equal(time.AddDurationToUnixNano(defaultTimeUnix, 42*stdtime.Second), invite.ExpiresTime)
	a.False(invite.IsExpired())

	time.TimeFunc = func() stdtime.Time {
		return defaultTime.Add(43 * stdtime.Second)
	}

	a.True(invite.IsExpired())
}

func TestUserAndInviteReadMethods(t *testing.T) {
	a := assert.New(t)

	b := BoardWithUsersAndInvites{
		Users: []BoardUser{
			{User: User{UserId: "u-123"}},
			{User: User{UserId: "u-456"}},
		},
		Invites: []BoardInvite{
			{InviteId: "i-123", User: User{UserId: "u-789"}},
		},
	}

	empty := BoardWithUsersAndInvites{}
	a.Equal(0, empty.UserCount())
	a.Equal(0, empty.InviteCount())

	a.False(empty.ContainsUser("u-123"))
	a.False(empty.ContainsInvite("i-123"))

	user, ok := empty.User("u-123")
	a.False(ok)
	a.Empty(user)

	a.Equal(2, b.UserCount())
	a.Equal(1, b.InviteCount())

	a.True(b.ContainsUser("u-123"))
	a.True(b.ContainsUser("u-456"))
	a.True(b.ContainsInvite("i-123"))
	a.True(b.ContainsInviteForUser("u-789"))

	user, ok = b.User("u-123")
	a.True(ok)
	a.Equal(BoardUser{User: User{UserId: "u-123"}}, user)

	invite, ok := b.Invite("i-123")
	a.True(ok)
	a.Equal(BoardInvite{InviteId: "i-123", User: User{UserId: "u-789"}}, invite)

	invite, ok = b.InviteForUser("u-789")
	a.True(ok)
	a.Equal(BoardInvite{InviteId: "i-123", User: User{UserId: "u-789"}}, invite)

	invite, ok = b.Invite("iii")
	a.False(ok)
	a.Empty(invite)
}
