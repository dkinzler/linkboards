package domain

import (
	"context"
	"testing"
	stdtime "time"

	"github.com/dkinzler/linkboards/internal/auth"

	"github.com/dkinzler/kit/errors"
	"github.com/dkinzler/kit/time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBoardDataStore struct {
	mock.Mock
}

func (m *MockBoardDataStore) UpdateBoard(ctx context.Context, boardId string, update *DatastoreBoardUpdate) error {
	args := m.Called(boardId, update)
	return args.Error(0)
}

func (m *MockBoardDataStore) DeleteBoard(ctx context.Context, boardId string) error {
	args := m.Called(boardId)
	return args.Error(0)
}

func (m *MockBoardDataStore) Board(ctx context.Context, boardId string) (BoardWithUsersAndInvites, TransactionExpectation, error) {
	args := m.Called(boardId)
	return args.Get(0).(BoardWithUsersAndInvites), args.Get(1), args.Error(2)
}

func (m *MockBoardDataStore) Boards(ctx context.Context, boardIds []string) ([]Board, error) {
	args := m.Called(boardIds)
	return args.Get(0).([]Board), args.Error(1)
}

func (m *MockBoardDataStore) BoardsForUser(ctx context.Context, userId string, qp QueryParams) ([]Board, error) {
	args := m.Called(userId, qp)
	return args.Get(0).([]Board), args.Error(1)
}

func (m *MockBoardDataStore) User(ctx context.Context, boardId string, userId string) (BoardUser, error) {
	args := m.Called(boardId, userId)
	return args.Get(0).(BoardUser), args.Error(1)
}

func (m *MockBoardDataStore) InvitesForUser(ctx context.Context, userId string, qp QueryParams) (map[string]BoardInvite, error) {
	args := m.Called(userId, qp)
	return args.Get(0).(map[string]BoardInvite), args.Error(1)
}

func newTestService() (BoardService, *MockBoardDataStore) {
	ds := &MockBoardDataStore{}
	return *NewBoardService(ds, nil), ds
}

var testTime = stdtime.Date(2022, 1, 1, 0, 0, 0, 0, stdtime.UTC)
var testTimeUnix = testTime.UnixNano()

func initTestTime() {
	time.TimeFunc = func() stdtime.Time {
		return testTime
	}
}

var user1 = User{UserId: "u-123"}
var user2 = User{UserId: "u-456"}
var user3 = User{UserId: "u-789"}
var user4 = User{UserId: "u-101"}

var exampleBoard = Board{
	BoardId:      "b-123",
	Name:         "test",
	Description:  "description",
	CreatedTime:  9001,
	CreatedBy:    user1,
	ModifiedTime: 9001,
	ModifiedBy:   user1,
}

var exampleBoardUser1 = BoardUser{
	User:         user1,
	Role:         auth.BoardRoleOwner,
	CreatedTime:  1000,
	InvitedBy:    user1,
	ModifiedTime: 1000,
	ModifiedBy:   user1,
}

var exampleBoardUser2 = BoardUser{
	User:         user2,
	Role:         auth.BoardRoleViewer,
	CreatedTime:  1000,
	InvitedBy:    user1,
	ModifiedTime: 1000,
	ModifiedBy:   user1,
}

var exampleBoardInvite1 = BoardInvite{
	InviteId:    "i-1",
	Role:        auth.BoardRoleViewer,
	CreatedTime: testTimeUnix,
	CreatedBy:   user1,
	ExpiresTime: testTimeUnix + 1000,
}

var exampleBoardInvite2 = BoardInvite{
	InviteId:    "i-2",
	Role:        auth.BoardRoleViewer,
	User:        user3,
	CreatedTime: testTimeUnix,
	CreatedBy:   user1,
	ExpiresTime: testTimeUnix + 1000,
}

var exampleBoardWithUAndI = BoardWithUsersAndInvites{
	Board:   exampleBoard,
	Users:   []BoardUser{exampleBoardUser1, exampleBoardUser2},
	Invites: []BoardInvite{exampleBoardInvite1, exampleBoardInvite2},
}

func TestCreateBoard(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	ctx := context.Background()
	initTestTime()

	// board name cannot be empty
	_, err := service.CreateBoard(ctx, "", "", User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errBoardNameEmpty))
	ds.AssertNotCalled(t, "UpdateBoard")

	// happy path, everything works
	ds.On("UpdateBoard", mock.Anything, mock.Anything).Return(nil).Once()
	board, err := service.CreateBoard(ctx, "abc", "description", User{UserId: "u-123"})
	a.Nil(err)
	a.True(len(board.Board.BoardId) > 10)
	a.Equal("abc", board.Board.Name)
	a.Equal("description", board.Board.Description)
	a.Equal(User{UserId: "u-123"}, board.Board.CreatedBy)
	a.Equal(User{UserId: "u-123"}, board.Board.ModifiedBy)
	a.Equal(testTimeUnix, board.Board.CreatedTime)
	a.Equal(testTimeUnix, board.Board.ModifiedTime)
	a.Len(board.Users, 1)
	boardUser := board.Users[0]
	a.Equal(auth.BoardRoleOwner, boardUser.Role)
	a.Equal("u-123", boardUser.User.UserId)
	ds.AssertExpectations(t)

	// fails when data store call fails
	ds.On("UpdateBoard", mock.Anything, mock.Anything).Return(errors.New(nil, "test", errors.Internal)).Once()
	board, err = service.CreateBoard(ctx, "abc", "description", User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.IsInternalError(err))
	a.Empty(board)
	ds.AssertExpectations(t)
}

func TestEditBoard(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	initTestTime()
	ctx := context.Background()

	// fails if board can't be returned
	ds.On("Board", "b-123").Return(BoardWithUsersAndInvites{}, nil, errors.New(nil, "test", errors.NotFound)).Once()
	_, err := service.EditBoard(ctx, "b-123", BoardEdit{
		UpdateName: true,
		Name:       "",
	}, User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.IsNotFoundError(err))
	ds.AssertExpectations(t)

	// fails, cannot change board name to empty
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	_, err = service.EditBoard(ctx, "b-123", BoardEdit{
		UpdateName: true,
		Name:       "",
	}, User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errBoardNameEmpty))
	ds.AssertExpectations(t)

	// fails if update operation fails
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.Anything).Return(errors.New(nil, "test", errors.Internal)).Once()
	_, err = service.EditBoard(ctx, "b-123", BoardEdit{
		UpdateName: true,
		Name:       "newName",
	}, User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.IsInternalError(err))
	ds.AssertExpectations(t)

	// happy path, this should work
	expected := exampleBoard
	expected.Name = "newName"
	expected.ModifiedTime = testTimeUnix
	expected.ModifiedBy = user4

	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.MatchedBy(func(update *DatastoreBoardUpdate) bool {
		if update == nil || !update.UpdateBoard {
			return false
		}
		if update.Board == expected {
			return true
		}
		return false
	})).Return(nil).Once()

	board, err := service.EditBoard(ctx, "b-123", BoardEdit{
		UpdateName: true,
		Name:       "newName",
	}, user4)
	a.Nil(err)
	a.Equal(expected, board)
	ds.AssertExpectations(t)
}

func TestCreateInvite(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	initTestTime()
	ctx := context.Background()

	// fails with invalid role
	_, err := service.CreateInvite(ctx, "b-123", "invalidrole", User{}, User{})
	a.NotNil(err)

	// fails if max number of invites reached
	invitesMax := make([]BoardInvite, maxInvitesPerBoard)
	ds.On("Board", "b-123").Return(BoardWithUsersAndInvites{Invites: invitesMax}, nil, nil).Once()
	_, err = service.CreateInvite(ctx, "b-123", auth.BoardRoleViewer, User{}, User{})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errMaxInvitesReached))
	ds.AssertExpectations(t)

	// fails if max number of invites reached
	usersMax := make([]BoardUser, maxUsersPerBoard)
	ds.On("Board", "b-123").Return(BoardWithUsersAndInvites{Users: usersMax}, nil, nil).Once()
	_, err = service.CreateInvite(ctx, "b-123", auth.BoardRoleViewer, User{}, User{})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errMaxBoardUsersReached))
	ds.AssertExpectations(t)

	// fails if user already has an invite
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	_, err = service.CreateInvite(ctx, "b-123", auth.BoardRoleViewer, user3, User{})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errUserAlreadyInvited))
	ds.AssertExpectations(t)

	// fails if user is already on the board
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	_, err = service.CreateInvite(ctx, "b-123", auth.BoardRoleViewer, user2, User{})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errUserAlreadyOnBoard))
	ds.AssertExpectations(t)

	//happy path, everything works
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.MatchedBy(func(update *DatastoreBoardUpdate) bool {
		if update == nil {
			return false
		}
		if len(update.UpdateInvites) != 1 {
			return false
		}
		bu := update.UpdateInvites[0]
		if bu.User != user4 {
			return false
		}
		if bu.Role != auth.BoardRoleViewer {
			return false
		}
		return true
	})).Return(nil).Once()
	_, err = service.CreateInvite(ctx, "b-123", auth.BoardRoleViewer, user4, User{})
	a.Nil(err)
	ds.AssertExpectations(t)
}

func TestDeleteInvite(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	initTestTime()
	ctx := context.Background()

	// fails if invite not found
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	err := service.DeleteInvite(ctx, "b-123", "i-3")
	a.NotNil(err)
	a.True(errors.IsNotFoundError(err))
	ds.AssertExpectations(t)

	// happy path
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.MatchedBy(func(update *DatastoreBoardUpdate) bool {
		if update == nil {
			return false
		}
		if len(update.RemoveInvites) == 1 {
			if update.RemoveInvites[0] == "i-1" {
				return true
			}
			return false
		}
		return false
	})).Return(nil).Once()
	err = service.DeleteInvite(ctx, "b-123", "i-1")
	a.Nil(err)
	ds.AssertExpectations(t)
}

func TestAcceptInvite(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	ctx := context.Background()

	time.TimeFunc = func() stdtime.Time {
		return stdtime.Unix(exampleBoardInvite1.ExpiresTime+1, 0)
	}

	// cannot accept expired invite
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	err := service.AcceptInvite(ctx, "b-123", "i-1", user4)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errInviteExpired))
	ds.AssertExpectations(t)

	initTestTime()

	// cannot accept invite for another user
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	err = service.AcceptInvite(ctx, "b-123", "i-2", user4)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errWrongUserForInvite))
	ds.AssertExpectations(t)

	// happy path, everything works
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.MatchedBy(func(update *DatastoreBoardUpdate) bool {
		if update == nil {
			return false
		}
		if !(len(update.RemoveInvites) == 1 && update.RemoveInvites[0] == "i-2") {
			return false
		}
		if !(len(update.UpdateUsers) == 1 && update.UpdateUsers[0].User == exampleBoardInvite2.User) {
			return false
		}
		return true
	})).Return(nil).Once()
	err = service.AcceptInvite(ctx, "b-123", "i-2", exampleBoardInvite2.User)
	a.Nil(err)
	ds.AssertExpectations(t)
}

func TestDeclineInvite(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	ctx := context.Background()
	initTestTime()

	// cannot decline invite not targeted at a user
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	err := service.DeclineInvite(ctx, "b-123", "i-1", user3)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errCannotDeclinePublicInvite))
	ds.AssertExpectations(t)

	// cannot decline invite for another user
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	err = service.DeclineInvite(ctx, "b-123", "i-2", user4)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errWrongUserForInvite))
	ds.AssertExpectations(t)

	// happy path, everything works
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.MatchedBy(func(update *DatastoreBoardUpdate) bool {
		if update == nil {
			return false
		}
		if !(len(update.RemoveInvites) == 1 && update.RemoveInvites[0] == "i-2") {
			return false
		}
		return true
	})).Return(nil).Once()
	err = service.DeclineInvite(ctx, "b-123", "i-2", user3)
	a.Nil(err)
	ds.AssertExpectations(t)
}

func TestRemoveUser(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	ctx := context.Background()
	initTestTime()

	// cannot remove board owner
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	err := service.RemoveUser(ctx, "b-123", exampleBoardWithUAndI.Board.CreatedBy.UserId)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errBoardOwnerCannotBeRemoved))
	ds.AssertExpectations(t)

	// cannot remove user not part of board
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	err = service.RemoveUser(ctx, "b-123", user4.UserId)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errUserNotOnBoard))
	ds.AssertExpectations(t)

	// happy path, everything works
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.MatchedBy(func(update *DatastoreBoardUpdate) bool {
		if update == nil {
			return false
		}
		if !(len(update.RemoveUsers) == 1 && update.RemoveUsers[0] == user2.UserId) {
			return false
		}
		return true
	})).Return(nil).Once()
	err = service.RemoveUser(ctx, "b-123", user2.UserId)
	a.Nil(err)
	ds.AssertExpectations(t)
}

func TestEditBoardUser(t *testing.T) {
	a := assert.New(t)
	service, ds := newTestService()
	ctx := context.Background()
	initTestTime()

	// cannot edit user not part of board
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	_, err := service.EditBoardUser(ctx, "b-123", user4.UserId, BoardUserEdit{UpdateRole: true, Role: auth.BoardRoleEditor}, user1)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errUserNotOnBoard))
	ds.AssertExpectations(t)

	// cannot change role to invalid one
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	_, err = service.EditBoardUser(ctx, "b-123", user2.UserId, BoardUserEdit{UpdateRole: true, Role: "invalidrole"}, user1)
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errInvalidRole))
	ds.AssertExpectations(t)

	// happy path, everything works
	ds.On("Board", "b-123").Return(exampleBoardWithUAndI, nil, nil).Once()
	ds.On("UpdateBoard", "b-123", mock.MatchedBy(func(update *DatastoreBoardUpdate) bool {
		if update == nil {
			return false
		}
		if len(update.UpdateUsers) == 0 {
			return false
		}
		bu := update.UpdateUsers[0]
		if bu.User != user2 {
			return false
		}
		if bu.Role != auth.BoardRoleEditor {
			return false
		}
		return true
	})).Return(nil).Once()
	_, err = service.EditBoardUser(ctx, "b-123", user2.UserId, BoardUserEdit{UpdateRole: true, Role: auth.BoardRoleEditor}, user1)
	a.Nil(err)
	ds.AssertExpectations(t)
}
