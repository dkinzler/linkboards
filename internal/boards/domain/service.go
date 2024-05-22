package domain

import (
	"context"
	stdtime "time"

	"github.com/dkinzler/linkboards/internal/auth"

	"github.com/dkinzler/kit/errors"
	"github.com/dkinzler/kit/time"
)

const (
	errBoardNameEmpty = iota + 1
	errBoardNameTooLong
	errBoardDescriptionTooLong
	errInvalidRole
	errMaxInvitesReached
	errMaxBoardUsersReached
	errUserAlreadyInvited
	errUserAlreadyOnBoard
	errInviteExpired
	errWrongUserForInvite
	errCannotDeclinePublicInvite
	errUserNotOnBoard
	errBoardOwnerCannotBeRemoved
	errOnlyCreatorCanBeOwner
	errCannotChangeRoleOfOwner
)

// BoardService provides operations on boards, users and invites.
// It uses an implementation of BoardDataStore to read and persist boards/users/invites,
// and an implementation of EventPublisher to make events available to other components/systems.
// BoardService does not perform authorization, that should be handled in application services using this package.
//
// On concurrency, consistency and transactions:
// We consider a board together with its users and invites as the unit of consistency,
// i.e. operations on BoardService can run concurrently only for different boards.
// If two operations e.g. try to concurrently create a new invite for the same user on the same board, at most one of the operations should succeed.
// BoardService guarantees consistency by using the optimistic transaction capabilities required of BoardDataStore implementations (see also the comments in datastore.go).
// In this application the number of users of a board is bounded, which in practice makes it very unlikely that concurrent operations on the same board,
// that could lead to inconsistencies, even happen that often.
//
// There are other ways of guaranteeing consistency, for example by making sure that
// requests for a given board are never processed concurrently. This would however also take some work, particularly
// if we want to run multiple instances of BoardService for scalability. Could e.g. route all requests with a given board id to the same instance.
// With optimistic transactions we can have many instances running at the same time
// without any additional coordination required while still maintaining consistency.
// (At least if the backing data store provides suitable consistency guarantees.)
type BoardService struct {
	ds BoardDataStore
	ep EventPublisher
}

// The BoardDataStore passed must not be nil.
// The EventPublisher can be, in which case the events will just end up nowhere.
func NewBoardService(ds BoardDataStore, ep EventPublisher) *BoardService {
	return &BoardService{ds: ds, ep: newMaybeEventPublisher(ep)}
}

func newServiceError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "domain/BoardService", code)
}

func (bs *BoardService) CreateBoard(ctx context.Context, name, description string, user User) (BoardWithUsersAndInvites, error) {
	board, err := NewBoard(name, description, user)
	if err != nil {
		return BoardWithUsersAndInvites{}, err
	}

	boardUser, err := NewBoardUser(user, auth.BoardRoleOwner, user)
	if err != nil {
		return BoardWithUsersAndInvites{}, err
	}

	err = bs.ds.UpdateBoard(ctx, board.BoardId, NewDatastoreBoardUpdate(nil).WithBoard(board).UpdateUser(boardUser))
	if err != nil {
		return BoardWithUsersAndInvites{}, newServiceError(err, errors.Internal).WithInternalMessage("could not create board")
	}

	bs.ep.PublishEvent(ctx, BoardCreated{
		BoardId:     board.BoardId,
		Name:        board.Name,
		Description: board.Description,
		CreatedTime: board.CreatedTime,
		CreatedBy:   board.CreatedBy,
	})

	return BoardWithUsersAndInvites{
		Board:   board,
		Users:   []BoardUser{boardUser},
		Invites: []BoardInvite{},
	}, nil
}

func (bs *BoardService) DeleteBoard(ctx context.Context, boardId string, user User) error {
	err := bs.ds.DeleteBoard(ctx, boardId)
	if err != nil {
		return err
	}

	bs.ep.PublishEvent(ctx, BoardDeleted{
		BoardId:   boardId,
		DeletedBy: user,
	})

	return nil
}

type BoardEdit struct {
	UpdateName        bool
	Name              string
	UpdateDescription bool
	Description       string
}

func (bs *BoardService) EditBoard(ctx context.Context, boardId string, be BoardEdit, user User) (Board, error) {
	b, te, err := bs.ds.Board(ctx, boardId)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return Board{}, newServiceError(err, errors.NotFound)
		}
		return Board{}, newServiceError(err, errors.Internal).WithInternalMessage("could not get invites")
	}

	board := b.Board

	if be.UpdateName {
		board.Name = be.Name
	}

	if be.UpdateDescription {
		board.Description = be.Description
	}

	err = board.IsValid()
	if err != nil {
		return Board{}, err
	}

	board.ModifiedTime = time.CurrTimeUnixNano()
	board.ModifiedBy = user

	err = bs.ds.UpdateBoard(ctx, boardId, NewDatastoreBoardUpdate(te).WithBoard(board))
	if err != nil {
		return Board{}, newServiceError(err, errors.Internal).WithInternalMessage("could not edit board")
	}

	return board, nil
}

const maxInvitesPerBoard = 32
const maxUsersPerBoard = 32
const inviteExpiryDuration = 3 * 24 * stdtime.Hour

func (bs *BoardService) CreateInvite(ctx context.Context, boardId string, role string, forUser User, fromUser User) (BoardInvite, error) {
	// Check first if invite is even valid, if not we don't even have to perform any data store calls
	invite, err := NewBoardInvite(role, fromUser, forUser, inviteExpiryDuration)
	if err != nil {
		return BoardInvite{}, err
	}

	board, te, err := bs.ds.Board(ctx, boardId)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return BoardInvite{}, newServiceError(err, errors.NotFound)
		}
		return BoardInvite{}, newServiceError(err, errors.Internal).WithInternalMessage("could not get board")
	}

	if board.InviteCount() >= maxInvitesPerBoard {
		return BoardInvite{}, newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("maximum number of invites reached").WithPublicCode(errMaxInvitesReached)
	}

	// cannot create an invite if board is full
	if board.UserCount() >= maxUsersPerBoard {
		return BoardInvite{}, newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("board already has maxiumum number of users").WithPublicCode(errMaxBoardUsersReached)
	}

	if forUser.UserId != "" {
		// Make sure user doesn't already have an invite
		if board.ContainsInviteForUser(forUser.UserId) {
			return BoardInvite{}, newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("user already invited").WithPublicCode(errUserAlreadyInvited)
		}

		// If the invite is for a specific user, the user should not already be part of the board.
		// We don't need to mind the transaction expectations here, since the user can only become part of the board with an invite and the invites are already "locked".
		// I.e. if the invites are modified while this request is running, the Update operation will fail anyway.
		if board.ContainsUser(forUser.UserId) {
			return BoardInvite{}, newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("user is already on board").WithPublicCode(errUserAlreadyOnBoard)
		}
	}

	err = bs.ds.UpdateBoard(ctx, boardId, NewDatastoreBoardUpdate(te).UpdateInvite(invite))
	if err != nil {
		return BoardInvite{}, newServiceError(err, errors.Internal).WithInternalMessage("could not update board")
	}

	return invite, nil
}

func (bs *BoardService) DeleteInvite(ctx context.Context, boardId string, inviteId string) error {
	board, te, err := bs.ds.Board(ctx, boardId)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return newServiceError(err, errors.NotFound)
		}
		return newServiceError(err, errors.Internal).WithInternalMessage("could not get board")
	}

	if !board.ContainsInvite(inviteId) {
		return newServiceError(nil, errors.NotFound)
	}

	err = bs.ds.UpdateBoard(ctx, boardId, NewDatastoreBoardUpdate(te).RemoveInvite(inviteId))
	if err != nil {
		return newServiceError(err, errors.Internal).WithInternalMessage("could not update board")
	}

	return nil
}

func (bs *BoardService) AcceptInvite(ctx context.Context, boardId string, inviteId string, user User) error {
	board, te, err := bs.ds.Board(ctx, boardId)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return newServiceError(err, errors.NotFound)
		}
		return newServiceError(err, errors.Internal).WithInternalMessage("could not get board")
	}

	invite, ok := board.Invite(inviteId)
	if !ok {
		return newServiceError(nil, errors.NotFound)
	}
	if invite.IsExpired() {
		return newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("invite expired").WithPublicCode(errInviteExpired)
	}

	if invite.User.UserId != "" {
		if invite.User.UserId != user.UserId {
			return newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("wrong user").WithPublicCode(errWrongUserForInvite)
		}
	}

	boardUser, err := NewBoardUser(user, invite.Role, invite.CreatedBy)
	if err != nil {
		return err
	}

	err = bs.ds.UpdateBoard(ctx, boardId, NewDatastoreBoardUpdate(te).RemoveInvite(inviteId).UpdateUser(boardUser))
	if err != nil {
		return newServiceError(err, errors.Internal).WithInternalMessage("could not update board")
	}

	return nil
}

// Only invites for a specific user can be declined.
func (bs *BoardService) DeclineInvite(ctx context.Context, boardId string, inviteId string, user User) error {
	board, te, err := bs.ds.Board(ctx, boardId)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return newServiceError(err, errors.NotFound)
		}
		return newServiceError(err, errors.Internal).WithInternalMessage("could not get board")
	}

	invite, ok := board.Invite(inviteId)
	if !ok {
		return newServiceError(nil, errors.NotFound)
	}

	if invite.User.UserId == "" {
		return newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("can only decline invites targeted at a specific user").WithPublicCode(errCannotDeclinePublicInvite)
	}
	if invite.User.UserId != user.UserId {
		return newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("wrong user").WithPublicCode(errWrongUserForInvite)
	}

	err = bs.ds.UpdateBoard(ctx, boardId, NewDatastoreBoardUpdate(te).RemoveInvite(inviteId))
	if err != nil {
		return newServiceError(err, errors.Internal).WithInternalMessage("could not update board")
	}

	return nil
}

func (bs *BoardService) RemoveUser(ctx context.Context, boardId string, userId string) error {
	board, te, err := bs.ds.Board(ctx, boardId)
	if err != nil {
		return newServiceError(err, errors.Internal).WithInternalMessage("could not get board")
	}

	if !board.ContainsUser(userId) {
		return newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("user not on board").WithPublicCode(errUserNotOnBoard)
	}

	user, _ := board.User(userId)
	if user.Role == auth.BoardRoleOwner {
		return newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("board owner cannot be removed").WithPublicCode(errBoardOwnerCannotBeRemoved)
	}

	err = bs.ds.UpdateBoard(ctx, boardId, NewDatastoreBoardUpdate(te).RemoveUser(userId))
	if err != nil {
		return newServiceError(err, errors.Internal).WithInternalMessage("could not update board")
	}

	return nil
}

type BoardUserEdit struct {
	UpdateRole bool
	Role       string
}

func (bs *BoardService) EditBoardUser(ctx context.Context, boardId string, userId string, bue BoardUserEdit, user User) (BoardUser, error) {
	board, te, err := bs.ds.Board(ctx, boardId)
	if err != nil {
		if errors.IsNotFoundError(err) {
			return BoardUser{}, newServiceError(err, errors.NotFound)
		}
		return BoardUser{}, newServiceError(err, errors.Internal).WithInternalMessage("could not get board")
	}

	if !board.ContainsUser(userId) {
		return BoardUser{}, newServiceError(nil, errors.FailedPrecondition).WithPublicMessage("user not on board").WithPublicCode(errUserNotOnBoard)
	}

	u, _ := board.User(userId)

	if bue.UpdateRole {
		err = u.ChangeRole(bue.Role, user)
		if err != nil {
			return BoardUser{}, err
		}
	}

	err = bs.ds.UpdateBoard(ctx, boardId, NewDatastoreBoardUpdate(te).UpdateUser(u))
	if err != nil {
		return BoardUser{}, newServiceError(err, errors.Internal).WithInternalMessage("could not update board")
	}

	return u, nil
}
