// Package inmem provides an in-memory implementation of BoardDataStore, that
// can be used for development/testing.
package inmem

import (
	"context"
	"sort"
	"sync"

	"github.com/d39b/linkboards/internal/boards/domain"

	"github.com/d39b/kit/errors"
)

type inmemBoardDataStore struct {
	m      sync.RWMutex
	boards map[string]domain.Board
	// maps from boardId to the set of users of that board
	// a set of users is represented as a map from userId to user
	users map[string]map[string]domain.BoardUser
	// maps from boardId to the set of invites of that board
	// a set of invites is represented as a map from inviteId to invite
	invites map[string]map[string]domain.BoardInvite
	// maps from boardId to version
	// the version of a board is increased every time it is modified
	version map[string]int
}

type transactionExpectation struct {
	boardId string
	version int
}

func NewInmemBoardDataStore() domain.BoardDataStore {
	return &inmemBoardDataStore{
		boards:  make(map[string]domain.Board),
		users:   make(map[string]map[string]domain.BoardUser),
		invites: make(map[string]map[string]domain.BoardInvite),
		version: make(map[string]int),
	}
}

func newError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "InmemBoardDataStore", code)
}

// should only be used from within another function that has acquired a read lock
func (s *inmemBoardDataStore) transactionExpectation(boardId string) domain.TransactionExpectation {
	version, ok := s.version[boardId]
	if !ok {
		return nil
	}
	return transactionExpectation{boardId: boardId, version: version}
}

func (s *inmemBoardDataStore) verifyTransactionExpectations(te domain.TransactionExpectation) bool {
	if te == nil {
		return true
	}
	t, ok := te.(transactionExpectation)
	if !ok {
		return false
	}
	version, ok := s.version[t.boardId]
	if !ok {
		// board no longer exists
		return false
	}
	if version != t.version {
		return false
	}
	return true
}

func (s *inmemBoardDataStore) UpdateBoard(ctx context.Context, boardId string, update *domain.DatastoreBoardUpdate) error {
	if update == nil || update.IsEmpty() {
		return newError(nil, errors.InvalidArgument).WithInternalMessage("empty update")
	}
	s.m.Lock()
	defer s.m.Unlock()
	ok := s.verifyTransactionExpectations(update.TransactionExpecation)
	if !ok {
		return newError(nil, errors.FailedPrecondition).WithInternalMessage("transaction expectations not met")
	}
	if update.UpdateBoard {
		s.boards[boardId] = update.Board
	}
	for _, user := range update.UpdateUsers {
		boardUsers, ok := s.users[boardId]
		if !ok {
			boardUsers = make(map[string]domain.BoardUser)
			s.users[boardId] = boardUsers
		}
		boardUsers[user.User.UserId] = user
	}
	for _, user := range update.RemoveUsers {
		boardUsers, ok := s.users[boardId]
		if ok {
			delete(boardUsers, user)
		}
	}
	for _, invite := range update.UpdateInvites {
		boardInvites, ok := s.invites[boardId]
		if !ok {
			boardInvites = make(map[string]domain.BoardInvite)
			s.invites[boardId] = boardInvites
		}
		boardInvites[invite.InviteId] = invite
	}
	for _, invite := range update.RemoveInvites {
		boardInvites, ok := s.invites[boardId]
		if ok {
			delete(boardInvites, invite)
		}
	}
	s.version[boardId] += 1
	return nil
}

func (s *inmemBoardDataStore) DeleteBoard(ctx context.Context, boardId string) error {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.boards, boardId)
	delete(s.users, boardId)
	delete(s.invites, boardId)
	delete(s.version, boardId)
	return nil
}

func (s *inmemBoardDataStore) Board(ctx context.Context, boardId string) (domain.BoardWithUsersAndInvites, domain.TransactionExpectation, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	board, ok := s.boards[boardId]
	if !ok {
		return domain.BoardWithUsersAndInvites{}, nil, newError(nil, errors.NotFound)
	}
	te := s.transactionExpectation(boardId)

	result := domain.BoardWithUsersAndInvites{}
	result.Board = board

	users, ok := s.users[boardId]
	if ok {
		result.Users = copyUsers(users)
	}
	invites, ok := s.invites[boardId]
	if ok {
		result.Invites = copyInvites(invites)
	}

	return result, te, nil
}

func (s *inmemBoardDataStore) Boards(ctx context.Context, boardIds []string) ([]domain.Board, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	result := make([]domain.Board, len(boardIds))
	for i, boardId := range boardIds {
		if board, ok := s.boards[boardId]; ok {
			result[i] = board
		}
	}

	return result, nil
}

type boardWithUserCreatedTime struct {
	Board       domain.Board
	CreatedTime int64
}

func (s *inmemBoardDataStore) BoardsForUser(ctx context.Context, userId string, qp domain.QueryParams) ([]domain.Board, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	boards := make([]boardWithUserCreatedTime, 0)
	for boardId, users := range s.users {
		if user, ok := users[userId]; ok {
			match := true
			if qp.Cursor != 0 {
				if user.CreatedTime > qp.Cursor {
					match = false
				}
			}

			if match {
				boards = append(boards, boardWithUserCreatedTime{
					Board:       s.boards[boardId],
					CreatedTime: user.CreatedTime,
				})
			}
		}
	}

	sort.Slice(boards, func(i, j int) bool {
		if boards[i].CreatedTime > boards[j].CreatedTime {
			return true
		}
		return false
	})

	if qp.Limit > 0 && qp.Limit < len(boards) {
		boards = boards[0:qp.Limit]
	}

	result := make([]domain.Board, len(boards))
	for i, board := range boards {
		result[i] = board.Board
	}

	return result, nil
}

func (s *inmemBoardDataStore) User(ctx context.Context, boardId string, userId string) (domain.BoardUser, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	users, ok := s.users[boardId]
	if !ok {
		return domain.BoardUser{}, newError(nil, errors.NotFound)
	}
	user, ok := users[userId]
	if !ok {
		return domain.BoardUser{}, newError(nil, errors.NotFound)
	}
	return user, nil
}

type inviteWithBoardId struct {
	boardId string
	invite  domain.BoardInvite
}

func (s *inmemBoardDataStore) InvitesForUser(ctx context.Context, userId string, qp domain.QueryParams) (map[string]domain.BoardInvite, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	result := make([]inviteWithBoardId, 0)
	for boardId, invites := range s.invites {
		for _, invite := range invites {
			if invite.User.UserId == userId {
				result = append(result, inviteWithBoardId{
					boardId: boardId,
					invite:  invite,
				})
			}
		}
	}

	if qp.Cursor != 0 {
		// filter results
		filtered := make([]inviteWithBoardId, 0, len(result))
		for _, invite := range result {
			if invite.invite.CreatedTime <= qp.Cursor {
				filtered = append(filtered, invite)
			}
		}
		result = filtered
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].invite.CreatedTime > result[j].invite.CreatedTime {
			return true
		}
		return false
	})

	if qp.Limit > 0 && qp.Limit < len(result) {
		result = result[0:qp.Limit]
	}

	resultMap := make(map[string]domain.BoardInvite, len(result))
	for _, invite := range result {
		resultMap[invite.boardId] = invite.invite
	}

	return resultMap, nil
}

func copyUsers(x map[string]domain.BoardUser) []domain.BoardUser {
	result := make([]domain.BoardUser, len(x))
	i := 0
	for _, user := range x {
		result[i] = user
		i++
	}
	return result
}

func copyInvites(x map[string]domain.BoardInvite) []domain.BoardInvite {
	result := make([]domain.BoardInvite, len(x))
	i := 0
	for _, invite := range x {
		result[i] = invite
		i++
	}
	return result
}
