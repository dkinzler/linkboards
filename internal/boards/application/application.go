package application

import (
	"context"
	"sort"

	"github.com/dkinzler/linkboards/internal/auth"
	"github.com/dkinzler/linkboards/internal/boards/domain"

	"github.com/dkinzler/kit/errors"
)

// @Kit{"endpointPackage":"internal/boards/transport", "httpPackage":"internal/boards/transport"}
type BoardApplicationService interface {
	// @Kit{
	//	"httpParams": ["json"],
	//	"endpoints": [{"http": {"path":"/boards", "method":"POST", "successCode": 201}}]
	// }
	CreateBoard(ctx context.Context, nb NewBoard) (BoardWithUsersAndInvites, error)
	// @Kit{
	//	"httpParams": ["url"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}", "method":"DELETE"}}]
	// }
	DeleteBoard(ctx context.Context, boardId string) error
	// @Kit{
	//	"httpParams": ["url", "json"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}", "method":"PATCH"}}]
	// }
	EditBoard(ctx context.Context, boardId string, be BoardEdit) (Board, error)
	// @Kit{
	//	"httpParams": ["url"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}", "method":"GET"}}]
	// }
	Board(ctx context.Context, boardId string) (BoardWithUsersAndInvites, error)
	// @Kit{
	//	"httpParams": ["query"],
	//	"endpoints": [{"http": {"path":"/boards", "method":"GET"}}]
	// }
	// Return boards the user making the request is part of.
	Boards(ctx context.Context, qp QueryParams) ([]Board, error)
	// @Kit{
	//	"httpParams": ["url", "json"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/invites", "method":"POST", "successCode": 201}}]
	// }
	CreateInvite(ctx context.Context, boardId string, ni NewInvite) (Invite, error)
	// @Kit{
	//	"httpParams": ["url", "url", "json"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/invites/{inviteId}", "method":"POST"}}]
	// }
	// Used to either accept or decline an invite
	RespondToInvite(ctx context.Context, boardId string, inviteId string, ir InviteResponse) error
	// @Kit{
	//	"httpParams": ["url", "url"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/invites/{inviteId}", "method":"DELETE"}}]
	// }
	DeleteInvite(ctx context.Context, boardId string, inviteId string) error
	// @Kit{
	//	"httpParams": ["query"],
	//	"endpoints": [{"http": {"path":"/invites", "method":"GET"}}]
	// }
	// Return invites for the user making the request.
	Invites(ctx context.Context, qp QueryParams) ([]Invite, error)
	// @Kit{
	//	"httpParams": ["url", "url"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/users/{userId}", "method":"DELETE"}}]
	// }
	RemoveUser(ctx context.Context, boardId string, userId string) error
	// @Kit{
	//	"httpParams": ["url", "url", "json"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/users/{userId}", "method":"PATCH"}}]
	// }
	EditBoardUser(ctx context.Context, boardId string, userId string, bue BoardUserEdit) (BoardUser, error)
	// We don't actually expose this method, it is there do demonstrate how we can use go concurrency
	// in service methods that assemble different pieces of data.
	BoardsAndInvites(ctx context.Context) (BoardsAndInvites, error)
}

type NewBoard struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Board struct {
	BoardId     string `json:"boardId"`
	Name        string `json:"name"`
	Description string `json:"description"`

	CreatedTime int64       `json:"createdTime"`
	CreatedBy   domain.User `json:"createdBy"`

	ModifiedTime int64       `json:"modifiedTime,omitempty"`
	ModifiedBy   domain.User `json:"modifiedBy,omitempty"`
}

type BoardEdit struct {
	// Use pointers here to differentiate between an empty value and value that was not provided.
	// E.g. assume the type of the name and description fields were string and we unmarshal a json http request body
	// into a BoardEdit value. The following to json bodies would yield the same BoardEdit value:
	//   - {"name":"abc"}, the user wants to change the name but leave the description field unchanged
	//   - {"name":"abc", "description":""}, the user wants to change the name but also set the description to the empty string
	// Both of these would yield a BoardEdit with the description field set to the empty string and we wouldn't know
	// if the user does not want to change the description or if they want to set it to the empty string.
	// With pointers we don't have this problem. They will be nil if the user didn't provide a value.
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (b BoardEdit) IsEmpty() bool {
	return b.Name == nil && b.Description == nil
}

func (b BoardEdit) ToDomainBoardEdit() domain.BoardEdit {
	var result domain.BoardEdit

	if b.Name != nil {
		result.UpdateName = true
		result.Name = *b.Name
	}

	if b.Description != nil {
		result.UpdateDescription = true
		result.Description = *b.Description
	}

	return result
}

type BoardWithUsersAndInvites struct {
	BoardId     string `json:"boardId"`
	Name        string `json:"name"`
	Description string `json:"description"`

	CreatedTime int64       `json:"createdTime"`
	CreatedBy   domain.User `json:"createdBy"`

	ModifiedTime int64       `json:"modifiedTime,omitempty"`
	ModifiedBy   domain.User `json:"modifiedBy,omitempty"`

	// should only be included if user has correct access
	Users   []BoardUser `json:"users"`
	Invites []Invite    `json:"invites"`
}

type BoardUser struct {
	User domain.User `json:"user"`
	Role string      `json:"role"`

	CreatedTime int64       `json:"createdTime"`
	InvitedBy   domain.User `json:"invitedBy"`

	ModifiedTime int64       `json:"modifiedTime"`
	ModifiedBy   domain.User `json:"modifiedBy"`
}

func userFromDomainUser(u domain.BoardUser) BoardUser {
	return BoardUser{
		User:         u.User,
		Role:         u.Role,
		CreatedTime:  u.CreatedTime,
		InvitedBy:    u.InvitedBy,
		ModifiedTime: u.ModifiedTime,
		ModifiedBy:   u.ModifiedBy,
	}
}

func usersFromDomainUsers(users []domain.BoardUser) []BoardUser {
	result := make([]BoardUser, len(users))
	for i, user := range users {
		result[i] = userFromDomainUser(user)
	}
	return result
}

type BoardUserEdit struct {
	Role *string `json:"role"`
}

func (b BoardUserEdit) IsEmpty() bool {
	return b.Role == nil
}

func (b BoardUserEdit) ToDomainBoardUserEdit() domain.BoardUserEdit {
	var result domain.BoardUserEdit
	if b.Role != nil {
		result.UpdateRole = true
		result.Role = *b.Role
	}
	return result
}

type NewInvite struct {
	// can be empty
	User domain.User `json:"user"`
	Role string      `json:"role"`
}

type Invite struct {
	BoardId  string `json:"boardId,omitempty"`
	InviteId string `json:"inviteId"`

	Role string      `json:"role"`
	User domain.User `json:"user"`

	CreatedTime int64       `json:"createdTime"`
	CreatedBy   domain.User `json:"createdBy"`

	ExpiresTime int64 `json:"expiresTime"`
}

const inviteResponseAccept = "accept"
const inviteResponseDecline = "decline"

type InviteResponse struct {
	// Accepted values are "accept" and "decline"
	Response string `json:"response"`
}

type QueryParams struct {
	// Max number of results to return
	Limit int
	// Cursor for pagination, only return results created at or before the given unix time (nanoseconds).
	// Can obtain a cursor value by e.g. taking the created time of the oldest element in the last query and then subtracting 1.
	Cursor int64
}

type BoardsAndInvites struct {
	Boards       []Board  `json:"boards,omitempty"`
	BoardsError  string   `json:"boardsError,omitempty"`
	Invites      []Invite `json:"invites,omitempty"`
	InvitesError string   `json:"invitesError,omitempty"`
}

type boardApplicationService struct {
	boardService   *domain.BoardService
	boardDataStore domain.BoardDataStore
	authChecker    *auth.BoardAuthorizationChecker
}

func NewBoardApplicationService(boardDataStore domain.BoardDataStore, authorizationStore auth.AuthorizationStore) BoardApplicationService {
	return &boardApplicationService{
		boardService:   domain.NewBoardService(boardDataStore, nil),
		boardDataStore: boardDataStore,
		authChecker:    NewAuthorizationChecker(authorizationStore),
	}
}

func newServiceError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "BoardApplicationService", code)
}

func newUnauthenticatedError() errors.Error {
	return newServiceError(nil, errors.Unauthenticated)
}

func newPermissionDeniedError() errors.Error {
	return newServiceError(nil, errors.PermissionDenied)
}

func (bas *boardApplicationService) CreateBoard(ctx context.Context, nb NewBoard) (BoardWithUsersAndInvites, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return BoardWithUsersAndInvites{}, newUnauthenticatedError()
	}

	// This isn't really necessary since right now the scopes an authenticated user has are fixed.
	// However in the future it might be possible to configure/change the scopes of authenticated users.
	if !authenticatedScopes.HasScope(createBoardScope) {
		return BoardWithUsersAndInvites{}, newPermissionDeniedError()
	}

	boardWithOwner, err := bas.boardService.CreateBoard(ctx, nb.Name, nb.Description, toDomainUser(user))
	if err != nil {
		return BoardWithUsersAndInvites{}, err
	}
	board := boardWithOwner.Board

	return BoardWithUsersAndInvites{
		BoardId:      board.BoardId,
		Name:         board.Name,
		Description:  board.Description,
		CreatedTime:  board.CreatedTime,
		CreatedBy:    board.CreatedBy,
		ModifiedTime: board.ModifiedTime,
		ModifiedBy:   board.ModifiedBy,
		Users:        usersFromDomainUsers(boardWithOwner.Users),
	}, nil
}

func (bas *boardApplicationService) DeleteBoard(ctx context.Context, boardId string) error {
	user, ok := userFromContext(ctx)
	if !ok {
		return newUnauthenticatedError()
	}

	az, err := bas.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return err
	}
	if !az.HasScope(deleteBoardScope) {
		return newPermissionDeniedError()
	}

	err = bas.boardService.DeleteBoard(ctx, boardId, toDomainUser(user))
	if err != nil {
		return err
	}

	return nil
}

func (bas *boardApplicationService) EditBoard(ctx context.Context, boardId string, be BoardEdit) (Board, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return Board{}, newUnauthenticatedError()
	}

	az, err := bas.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return Board{}, err
	}
	if !az.HasScope(editBoardScope) {
		return Board{}, newPermissionDeniedError()
	}

	if be.IsEmpty() {
		return Board{}, newServiceError(nil, errors.InvalidArgument).WithPublicMessage("empty update")
	}

	board, err := bas.boardService.EditBoard(ctx, boardId, be.ToDomainBoardEdit(), toDomainUser(user))
	if err != nil {
		return Board{}, err
	}

	return Board{
		BoardId:      boardId,
		Name:         board.Name,
		Description:  board.Description,
		CreatedTime:  board.CreatedTime,
		CreatedBy:    board.CreatedBy,
		ModifiedTime: board.ModifiedTime,
		ModifiedBy:   board.ModifiedBy,
	}, nil
}

func (bas *boardApplicationService) Board(ctx context.Context, boardId string) (BoardWithUsersAndInvites, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return BoardWithUsersAndInvites{}, newUnauthenticatedError()
	}

	az, err := bas.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return BoardWithUsersAndInvites{}, err
	}
	if !az.HasScope(viewBoardScope) {
		return BoardWithUsersAndInvites{}, newPermissionDeniedError()
	}

	b, _, err := bas.boardDataStore.Board(ctx, boardId)
	if err != nil {
		return BoardWithUsersAndInvites{}, err
	}

	board := b.Board

	result := BoardWithUsersAndInvites{
		BoardId:     board.BoardId,
		Name:        board.Name,
		Description: board.Description,
		CreatedTime: board.CreatedTime,
		CreatedBy:   board.CreatedBy,
	}

	if az.HasScope(editBoardScope) {
		result.ModifiedTime = board.ModifiedTime
		result.ModifiedBy = board.ModifiedBy
	}

	if az.HasScope(viewBoardUsersScope) {
		boardUsers := make([]BoardUser, len(b.Users))
		for i, u := range b.Users {
			boardUsers[i] = BoardUser{
				User:         u.User,
				Role:         u.Role,
				CreatedTime:  u.CreatedTime,
				InvitedBy:    u.InvitedBy,
				ModifiedTime: u.ModifiedTime,
				ModifiedBy:   u.ModifiedBy,
			}
		}
		result.Users = boardUsers
	}

	if az.HasScope(viewBoardInvitesScope) {
		boardInvites := make([]Invite, len(b.Invites))
		for i, inv := range b.Invites {
			boardInvites[i] = Invite{
				InviteId:    inv.InviteId,
				Role:        inv.Role,
				User:        inv.User,
				CreatedTime: inv.CreatedTime,
				CreatedBy:   inv.CreatedBy,
				ExpiresTime: inv.ExpiresTime,
			}
		}
		result.Invites = boardInvites
	}

	return result, nil
}

func (bas *boardApplicationService) Boards(ctx context.Context, qp QueryParams) ([]Board, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return nil, newUnauthenticatedError()
	}

	if !authenticatedScopes.HasScope(listUserBoardsScope) {
		return nil, newPermissionDeniedError()
	}

	dqp := domain.NewQueryParams()
	if qp.Limit != 0 {
		dqp = dqp.WithLimit(qp.Limit)
	}
	if qp.Cursor != 0 {
		dqp = dqp.WithCursor(qp.Cursor)
	}

	boards, err := bas.boardDataStore.BoardsForUser(ctx, user.UserId, dqp)
	if err != nil {
		return nil, err
	}

	result := make([]Board, len(boards))
	for i, b := range boards {
		result[i] = Board{
			BoardId:     b.BoardId,
			Name:        b.Name,
			Description: b.Description,
			CreatedTime: b.CreatedTime,
			CreatedBy:   b.CreatedBy,
		}
	}
	return result, nil
}

func (bas *boardApplicationService) CreateInvite(ctx context.Context, boardId string, ni NewInvite) (Invite, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return Invite{}, newUnauthenticatedError()
	}

	az, err := bas.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return Invite{}, err
	}
	if !az.HasScope(createInviteScope) {
		return Invite{}, newPermissionDeniedError()
	}

	invite, err := bas.boardService.CreateInvite(ctx, boardId, ni.Role, ni.User, toDomainUser(user))
	if err != nil {
		return Invite{}, err
	}

	return Invite{
		BoardId:     boardId,
		InviteId:    invite.InviteId,
		Role:        invite.Role,
		User:        invite.User,
		CreatedTime: invite.CreatedTime,
		CreatedBy:   invite.CreatedBy,
		ExpiresTime: invite.ExpiresTime,
	}, nil
}

func (bas *boardApplicationService) RespondToInvite(ctx context.Context, boardId string, inviteId string, r InviteResponse) error {
	user, ok := userFromContext(ctx)
	if !ok {
		return newUnauthenticatedError()
	}

	if !authenticatedScopes.HasScope(respondToInviteScope) {
		return newPermissionDeniedError()
	}

	if r.Response == inviteResponseAccept {
		return bas.boardService.AcceptInvite(ctx, boardId, inviteId, toDomainUser(user))
	} else if r.Response == inviteResponseDecline {
		return bas.boardService.DeclineInvite(ctx, boardId, inviteId, toDomainUser(user))
	} else {
		return newServiceError(nil, errors.InvalidArgument).WithPublicMessage("invalid invite response")
	}
}

func (bas *boardApplicationService) DeleteInvite(ctx context.Context, boardId string, inviteId string) error {
	user, ok := userFromContext(ctx)
	if !ok {
		return newUnauthenticatedError()
	}

	az, err := bas.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return err
	}
	if !az.HasScope(deleteInviteScope) {
		return newPermissionDeniedError()
	}

	err = bas.boardService.DeleteInvite(ctx, boardId, inviteId)
	if err != nil {
		return err
	}

	return nil
}

func (bas *boardApplicationService) Invites(ctx context.Context, qp QueryParams) ([]Invite, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return nil, newUnauthenticatedError()
	}

	if !authenticatedScopes.HasScope(listUserInvitesScope) {
		return nil, newPermissionDeniedError()
	}

	dqp := domain.NewQueryParams()
	if qp.Limit != 0 {
		dqp = dqp.WithLimit(qp.Limit)
	}
	if qp.Cursor != 0 {
		dqp = dqp.WithCursor(qp.Cursor)
	}

	invites, err := bas.boardDataStore.InvitesForUser(ctx, user.UserId, dqp)
	if err != nil {
		return nil, err
	}

	result := make([]Invite, 0, len(invites))
	for boardId, invite := range invites {
		result = append(result, Invite{
			BoardId:     boardId,
			InviteId:    invite.InviteId,
			Role:        invite.Role,
			User:        invite.User,
			CreatedTime: invite.CreatedTime,
			CreatedBy:   invite.CreatedBy,
			ExpiresTime: invite.ExpiresTime,
		})
	}

	// sort result from newest to oldest
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedTime > result[j].CreatedTime {
			return true
		}
		return false
	})

	return result, nil
}

func (bas *boardApplicationService) RemoveUser(ctx context.Context, boardId string, userId string) error {
	user, ok := userFromContext(ctx)
	if !ok {
		return newUnauthenticatedError()
	}

	// Users can remove themselves, i.e. leave a board
	// Otherwise a user needs the correct permission
	if user.UserId != userId {
		az, err := bas.authChecker.GetAuthorization(ctx, boardId, user.UserId)
		if err != nil {
			return err
		}

		if !az.HasScope(removeUserFromBoardScope) {
			return newPermissionDeniedError()
		}
	}

	err := bas.boardService.RemoveUser(ctx, boardId, userId)
	if err != nil {
		return err
	}

	return nil
}

func (bas *boardApplicationService) EditBoardUser(ctx context.Context, boardId string, userId string, bue BoardUserEdit) (BoardUser, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return BoardUser{}, newUnauthenticatedError()
	}

	az, err := bas.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return BoardUser{}, err
	}
	if !az.HasScope(editBoardUserScope) {
		return BoardUser{}, newPermissionDeniedError()
	}

	if bue.IsEmpty() {
		return BoardUser{}, newServiceError(nil, errors.InvalidArgument).WithPublicMessage("empty update")
	}

	boardUser, err := bas.boardService.EditBoardUser(ctx, boardId, userId, bue.ToDomainBoardUserEdit(), toDomainUser(user))
	if err != nil {
		return BoardUser{}, err
	}

	return BoardUser{
		User:         boardUser.User,
		Role:         boardUser.Role,
		CreatedTime:  boardUser.CreatedTime,
		InvitedBy:    boardUser.InvitedBy,
		ModifiedTime: boardUser.ModifiedTime,
		ModifiedBy:   boardUser.ModifiedBy,
	}, nil
}

// We don't actually expose this method, it is there do demonstrate how we can use go concurrency
// in service methods that assemble different pieces of data.
// Will return both boards and invites for the user making the request.
// For some clients it might make more sense to retrieve the data together in one request instead of performing multiple.
func (bas *boardApplicationService) BoardsAndInvites(ctx context.Context) (BoardsAndInvites, error) {
	user, ok := userFromContext(ctx)
	if !ok {
		return BoardsAndInvites{}, newUnauthenticatedError()
	}

	if !(authenticatedScopes.HasScope(listUserBoardsScope) && authenticatedScopes.HasScope(listUserInvitesScope)) {
		return BoardsAndInvites{}, newPermissionDeniedError()
	}

	// The following will start two go routines to load the boards and invites concurrently.
	// If at least one of those operations succeed, we return the data that was available, so that client get at least something.

	bc := runConcurrent(func() ([]Board, error) {
		boards, err := bas.boardDataStore.BoardsForUser(ctx, user.UserId, domain.NewQueryParams())
		if err != nil {
			return nil, err
		}
		result := make([]Board, len(boards))
		for i, b := range boards {
			result[i] = Board{
				BoardId:     b.BoardId,
				Name:        b.Name,
				Description: b.Description,
				CreatedTime: b.CreatedTime,
				CreatedBy:   b.CreatedBy,
			}
		}
		return result, nil
	})
	ic := runConcurrent(func() ([]Invite, error) {
		invites, err := bas.boardDataStore.InvitesForUser(ctx, user.UserId, domain.NewQueryParams())
		if err != nil {
			return nil, err
		}
		result := make([]Invite, 0, len(invites))
		for boardId, invite := range invites {
			result = append(result, Invite{
				BoardId:     boardId,
				InviteId:    invite.InviteId,
				Role:        invite.Role,
				User:        invite.User,
				CreatedTime: invite.CreatedTime,
				CreatedBy:   invite.CreatedBy,
				ExpiresTime: invite.ExpiresTime,
			})
		}
		return result, nil
	})

	result := BoardsAndInvites{}
	for i := 0; i < 2; i++ {
		select {
		case boards := <-bc:
			if boards.err == nil {
				result.Boards = boards.result
			} else {
				result.BoardsError = "Could not load boards"
			}
		case invites := <-ic:
			if invites.err == nil {
				result.Invites = invites.result
			} else {
				result.InvitesError = "Could not load invites"
			}
		}
	}

	if result.BoardsError != "" && result.InvitesError != "" {
		return BoardsAndInvites{}, newServiceError(nil, errors.Internal)
	}

	return result, nil
}

type result[T any] struct {
	result T
	err    error
}

// runs the function in a new go routine and returns a channel on which the
// result of the function will be sent once it's done
func runConcurrent[T any](f func() (T, error)) <-chan result[T] {
	rChan := make(chan result[T], 1)
	go func() {
		t, err := f()
		rChan <- result[T]{
			result: t,
			err:    err,
		}
	}()
	return rChan
}
