// Package domain implements the business logic for working with boards, users and invites:
//   - functions to create, validate and work with instances of boards/users/invites
//   - general data store interface BoardDataStore, which makes it easy to switch out the actual storage mechanism for the data
//   - BoardService implements operations that can be performed on boards, users and invites
package domain

import (
	"linkboards/internal/auth"
	stdtime "time"

	"github.com/d39b/kit/errors"
	"github.com/d39b/kit/time"
	"github.com/d39b/kit/uuid"
)

// User represents the information about a user we need in this context.
// A User value is used in other types to e.g. represent the user that created a board or an invite.
//
// We assume that users/authentication are managed elsewhere, e.g. another package or an outside system like Firebase Authentication.
// There a user might have more attributes like an email address, a photo, or birth date.
//
// While the id of a user shouldn't change, the user can possibly change their name in the authentication system that manages users.
type User struct {
	UserId string
	Name   string
}

type Board struct {
	// Random UUIDv4 with prefix "b-"
	BoardId string

	Name        string
	Description string

	// Time the board was created as Unix time (nanoseconds)
	CreatedTime int64
	// The user that created this board.
	// Note that the User value contains the id and name of the user.
	// While the id should be unique and never change, a user usually has the ability to change their name.
	// The name of the user kept with this board would then no longer be up to date.
	// If we don't want to accept these inconsistencies, we would need to store just the user id and every time a user's name is required
	// we would have to read the latest value from the authentication system manages users and their names.
	CreatedBy User
	// Last time the board was modified (e.g. the name or description changed) as Unix time (nanoseconds)
	ModifiedTime int64
	// User that performed the latest modification
	ModifiedBy User
}

func newError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "boards/domain", code)
}

func NewBoard(name, description string, user User) (Board, error) {
	id, err := uuid.NewUUIDWithPrefix("b")
	if err != nil {
		return Board{}, newError(err, errors.Internal).WithInternalMessage("error creating random uuid")
	}

	timeNow := time.CurrTimeUnixNano()

	board := Board{
		BoardId:      id,
		Name:         name,
		Description:  description,
		CreatedTime:  timeNow,
		CreatedBy:    user,
		ModifiedTime: timeNow,
		ModifiedBy:   user,
	}

	err = board.IsValid()
	if err != nil {
		return Board{}, err
	}

	return board, nil
}

func (b *Board) IsValid() error {
	err := b.IsNameValid()
	if err != nil {
		return err
	}
	return b.IsDescriptionValid()
}

const nameMaxLength = 100

// A board name is valid if it is not empty and not longer than nameMaxLength bytes.
func (b *Board) IsNameValid() error {
	if len(b.Name) == 0 {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("empty name").WithPublicCode(errBoardNameEmpty)
	}
	if len(b.Name) > nameMaxLength {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("name too long").WithPublicCode(errBoardNameTooLong)
	}
	return nil
}

const descriptionMaxLength = 1000

// A board description can be at most descriptionMaxLength bytes long.
func (b *Board) IsDescriptionValid() error {
	if len(b.Description) > descriptionMaxLength {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("description too long").WithPublicCode(errBoardDescriptionTooLong)
	}
	return nil
}

type BoardUser struct {
	User User
	// Role the user has for the board, see the auth package for available roles.
	// Determines what the user can do on the board.
	Role string
	// Time the user joined the board as Unix time (nanoseconds).
	CreatedTime int64
	// The user that invited this user to the board.
	InvitedBy User
	// Last time this BoardUser was modified (e.g. the role changed) as Unix time (nanoseconds).
	ModifiedTime int64
	ModifiedBy   User
}

func NewBoardUser(user User, role string, invitedBy User) (BoardUser, error) {
	if user.UserId == "" {
		return BoardUser{}, newError(nil, errors.InvalidArgument).WithInternalMessage("user id empty")
	}

	if !auth.IsBoardRoleValid(role) {
		return BoardUser{}, newError(nil, errors.InvalidArgument).WithPublicMessage("invalid role").WithPublicCode(errInvalidRole)
	}

	timeNow := time.CurrTimeUnixNano()

	return BoardUser{
		User:         user,
		Role:         role,
		CreatedTime:  timeNow,
		InvitedBy:    invitedBy,
		ModifiedTime: timeNow,
		ModifiedBy:   invitedBy,
	}, nil
}

func (b *BoardUser) ChangeRole(role string, modifiedBy User) error {
	if !auth.IsBoardRoleValid(role) {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("invalid role").WithPublicCode(errInvalidRole)
	}
	// If user has owner role, their role cannot be changed.
	if b.Role == auth.BoardRoleOwner {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("cannot change the role of the board owner").WithPublicCode(errCannotChangeRoleOfOwner)
	}
	// Cannot change a users role to owner, only creator of a board can be owner.
	if role == auth.BoardRoleOwner {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("only creator can be owner").WithPublicCode(errOnlyCreatorCanBeOwner)
	}
	b.Role = role
	b.ModifiedTime = time.CurrTimeUnixNano()
	b.ModifiedBy = modifiedBy
	return nil
}

// An invite to join a board.
// Can be either a general invite, that can be accepted by any user or an invite specific to a user.
type BoardInvite struct {
	// Random UUIDv4 with prefix "i-"
	InviteId string

	// Role a user that accepts the invite would have on the board.
	Role string
	// If not empty, only the given user can accept the invite.
	// If empty, any user can.
	User User

	// Time the invite was created as Unix time (nanoseconds).
	CreatedTime int64
	// User that created the invite.
	CreatedBy User

	// Time the invite expires as Unix time (nanoseconds).
	// After this time the invite can no longer be accepted.
	ExpiresTime int64
}

func NewBoardInvite(role string, createdBy User, forUser User, expiryDuration stdtime.Duration) (BoardInvite, error) {
	if !auth.IsBoardRoleValid(role) {
		return BoardInvite{}, newError(nil, errors.InvalidArgument).WithPublicMessage("invalid role").WithPublicCode(errInvalidRole)
	}

	// A board can have only one user with the owner role and that is the creator of the board.
	// If this is changed in the future, make sure that no privilege escalation is possible.
	if role == auth.BoardRoleOwner {
		return BoardInvite{}, newServiceError(nil, errors.InvalidArgument).WithPublicMessage("cannot invite user with owner role").WithPublicCode(errOnlyCreatorCanBeOwner)
	}

	id, err := uuid.NewUUIDWithPrefix("i")
	if err != nil {
		return BoardInvite{}, newError(nil, errors.Internal).WithInternalMessage("error creating uuid")
	}

	currTime := time.CurrTimeUnixNano()
	expiresTime := time.AddDurationToUnixNano(currTime, expiryDuration)

	return BoardInvite{
		InviteId:    id,
		Role:        role,
		User:        forUser,
		CreatedTime: currTime,
		CreatedBy:   createdBy,
		ExpiresTime: expiresTime,
	}, nil
}

// Returns true if the invite has expired.
func (i BoardInvite) IsExpired() bool {
	expiresTime := stdtime.Unix(0, i.ExpiresTime)
	return time.CurrTime().After(expiresTime)
}

// A board with all its users and invites, which we consider
// the unit of consistency. I.e. the service methods supported by
// the datastore should guarantee that there are no
// concurrent operations on a single board (including its users and invites).
//
// This struct does not provide methods to manipulate the sets of users and invites,
// instead the updates should be encoded using a DatastoreBoardUpdate value.
//
// Note also that since the number of users and invites for a single board is limited,
// we do not have to worry about the implications of loading all users/invites for performance and memory.
type BoardWithUsersAndInvites struct {
	Board   Board
	Users   []BoardUser
	Invites []BoardInvite
}

func (b BoardWithUsersAndInvites) ContainsInvite(inviteId string) bool {
	for _, invite := range b.Invites {
		if invite.InviteId == inviteId {
			return true
		}
	}
	return false
}

func (b BoardWithUsersAndInvites) ContainsInviteForUser(userId string) bool {
	for _, invite := range b.Invites {
		if invite.User.UserId == userId {
			return true
		}
	}
	return false
}

func (b BoardWithUsersAndInvites) Invite(inviteId string) (BoardInvite, bool) {
	for _, invite := range b.Invites {
		if invite.InviteId == inviteId {
			return invite, true
		}
	}
	return BoardInvite{}, false
}

func (b BoardWithUsersAndInvites) InviteForUser(userId string) (BoardInvite, bool) {
	for _, invite := range b.Invites {
		if invite.User.UserId == userId {
			return invite, true
		}
	}
	return BoardInvite{}, false
}

func (b BoardWithUsersAndInvites) InviteCount() int {
	return len(b.Invites)
}

func (b BoardWithUsersAndInvites) ContainsUser(userId string) bool {
	for _, user := range b.Users {
		if user.User.UserId == userId {
			return true
		}
	}
	return false
}

func (b BoardWithUsersAndInvites) User(userId string) (BoardUser, bool) {
	for _, user := range b.Users {
		if user.User.UserId == userId {
			return user, true
		}
	}
	return BoardUser{}, false
}

func (b BoardWithUsersAndInvites) UserCount() int {
	return len(b.Users)
}
