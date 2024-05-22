package domain

import "context"

// Any type implementing this interface can be used as the data store for boards, users and invites.
// We consider a board together with its users and invites as the unit of consistency, they can be retrieved
// and modified together.
// To avoid inconsistencies when multiple concurrent requests try to modify a board (and/or its users/invites),
// we use optimistic transactions/concurrency control (see also the comments on the TransactionExpectation interface and the Board() and UpdateBoard() methods).
type BoardDataStore interface {
	// Update a single board and/or some users/invites (belonging to a single board).
	//
	// To implement optimistic concurrency, this method takes a TransactionExpectation (as part of the update parameter).
	// The TransactionExpectation represents the last time the board was modified.
	// Any implementation of this interface should guarantee that the board (and/or its users/invites) is updated only
	// if the board wasn't changed (i.e. it is still in the same state represented by the TransactionExpectation).
	UpdateBoard(ctx context.Context, boardId string, update *DatastoreBoardUpdate) error
	DeleteBoard(ctx context.Context, boardId string) error
	// Returns the board with the given id together with all its users and invites.
	// There are (almost) no separate methods to query or retrieve specific users/invites.
	// This is feasible since the number of users and invites per board is limited.
	// If in the future we wanted to remove these limits, it would make sense to treat boards, board users and board invites
	// as separate entities, with their own data store methods to retrieve and modify them.
	//
	// This method returns a TransactionExpectation that can later be passed back into the UpdateBoard method,
	// to make sure that the board (and its users/invites) was not changed.
	Board(ctx context.Context, boardId string) (BoardWithUsersAndInvites, TransactionExpectation, error)
	Boards(ctx context.Context, boardIds []string) ([]Board, error)
	BoardsForUser(ctx context.Context, userId string, qp QueryParams) ([]Board, error)

	User(ctx context.Context, boardId string, userId string) (BoardUser, error)

	InvitesForUser(ctx context.Context, userId string, qp QueryParams) (map[string]BoardInvite, error)
}

// Defines an update to a single board and its users and/or invites.
//
// The set of changes to the users/invites are defined using update and remove lists, instead of just
// passing the new complete lists of users/invites.
// This might be useful for data store implementations that keep users and invites separately from their board,
// e.g. to optimize or make it easier to perform certain queries.
type DatastoreBoardUpdate struct {
	TransactionExpecation TransactionExpectation

	// Set to true if the board should be updated with the value defined by the "Board" field.
	UpdateBoard bool
	Board       Board

	UpdateUsers []BoardUser
	// User ids of users that should be removed
	RemoveUsers []string

	UpdateInvites []BoardInvite
	// Invite ids of invites that should be removed
	RemoveInvites []string
}

func NewDatastoreBoardUpdate(te TransactionExpectation) *DatastoreBoardUpdate {
	return &DatastoreBoardUpdate{TransactionExpecation: te}
}

func (u *DatastoreBoardUpdate) WithBoard(b Board) *DatastoreBoardUpdate {
	u.UpdateBoard = true
	u.Board = b
	return u
}

func (u *DatastoreBoardUpdate) UpdateUser(user BoardUser) *DatastoreBoardUpdate {
	u.UpdateUsers = append(u.UpdateUsers, user)
	return u
}

func (u *DatastoreBoardUpdate) RemoveUser(userId string) *DatastoreBoardUpdate {
	u.RemoveUsers = append(u.RemoveUsers, userId)
	return u
}

func (u *DatastoreBoardUpdate) UpdateInvite(invite BoardInvite) *DatastoreBoardUpdate {
	u.UpdateInvites = append(u.UpdateInvites, invite)
	return u
}

func (u *DatastoreBoardUpdate) RemoveInvite(inviteId string) *DatastoreBoardUpdate {
	u.RemoveInvites = append(u.RemoveInvites, inviteId)
	return u
}

func (u *DatastoreBoardUpdate) IsEmpty() bool {
	return !(u.UpdateBoard || len(u.UpdateUsers) > 0 || len(u.RemoveUsers) > 0 || len(u.UpdateInvites) > 0 || len(u.RemoveInvites) > 0)
}

// A TransactionExpectation should contain information about when some piece of data
// was last modified.
// It is used to implement optimistic transactions/concurrency control.
//
// In particular for BoardDataStore, TransactionExpectation can be used to guarantee consistency in the following way:
//   - We can retrieve a board (with all the users and invites) using the Board() method.
//   - Board() also returns a value implementing TransactionExpectation, it will typically contain the last time the
//     board (and/or its users/invites) was modified.
//   - When we want to update/modify the board (and/or its users/invites), we can pass the TransactionExpectation back to the
//     UpdateBoard() method (as part of the "update" parameter).
//   - A correct data store implementation will have to guarantee that the update/modification is only performed
//     if the board (and/or users/invites) was not modified since the time defined in the TransactionExpectation value.
//
// Right now, we will only ever need to pass a single TransactionExpectation.
// However for a more complicated scenario, the TransactionExpectation interface could require a method
// "Combine(other TransactionExpectation) TransactionExpectation" that can be used to combine multiple expectations.
// I.e. the combined expectation would be used to represent the last modification time of multiple pieces of data.
//
// There are other ways of implementing optimistic transactions, for example:
//   - Instead of explicitly passing back and forth TransactionExpectation values, BoardDataStore could store and retrieve those from the context.Context.
//     A drawback of this approach is that even more of the responsibility for assuring consistency is put on BoardDataStore implementations.
//     It is no longer easy to see by looking at the BoardService code what consistency is guaranteed.
//   - Pass an update function "func(oldValue Board) (newValue Board) { ... }" to BoardDataStore.
//     The data store implementation can use this function in whatever way necessary but has to guarantee
//     that the update happens consistently.
type TransactionExpectation interface{}

// Parameters for data store queries for boards and invites.
// Can be used to limit the number of results returned
// or provide a cursor for pagination.
// By default, elements will be sorted from newest to oldest (i.e. by descending created time).
type QueryParams struct {
	// Valid limit values are between 1 and 100 (inclusive)
	Limit int
	// Cursor for pagination, only return elements that were created at or before the given Unix time (nanoseconds).
	// To obtain the cursor value, one can use the created time of the oldest element in the last query and subtract 1.
	// Note that this could skip some elements, if they were created at the exact same time, however this is very unlikely
	// since we use Unix time in nanoseconds.
	Cursor int64
}

func NewQueryParams() QueryParams {
	return QueryParams{Limit: 20}
}

func (qp QueryParams) WithLimit(limit int) QueryParams {
	if limit >= 1 && limit <= 100 {
		qp.Limit = limit
	}
	return qp
}

func (qp QueryParams) WithCursor(cursor int64) QueryParams {
	if cursor > 0 {
		qp.Cursor = cursor
	}
	return qp
}
