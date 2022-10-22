package firestore

import (
	"context"
	"linkboards/internal/boards/domain"

	"github.com/d39b/kit/errors"
	fs "github.com/d39b/kit/firebase/firestore"

	"cloud.google.com/go/firestore"
)

const boardCollectionName = "boards"
const boardUserCollectionName = "boardUsers"
const boardInviteCollectionName = "boardInvites"

// We store a board and all its users and invites together in a single document,
// since we often want to read all that data together.
// The number of users and invites is limited, so we don't have to worry about a document
// becoming too large (firestore 1MB document size limit).
// To make it easier to query and sort all the boards a user is a member of and all the invites for a user (across boards),
// we also store board users and invites as separate documents in separate collections.
//
// The cost of this approach is that we have to perform multiple writes when updating a board or adding/removing users/invites.
// However the number of times the board data is read is probably much higher than the number of changes to the board and its users/invites.
type firestoreBoardDataStore struct {
	client                *firestore.Client
	boardCollection       *firestore.CollectionRef
	boardUserCollection   *firestore.CollectionRef
	boardInviteCollection *firestore.CollectionRef
}

func NewFirestoreBoardDataStore(client *firestore.Client) domain.BoardDataStore {
	return &firestoreBoardDataStore{
		client:                client,
		boardCollection:       client.Collection(boardCollectionName),
		boardUserCollection:   client.Collection(boardUserCollectionName),
		boardInviteCollection: client.Collection(boardInviteCollectionName),
	}
}

type transactionExpectation fs.TransactionExpectations

func (te transactionExpectation) Combine(other domain.TransactionExpectation) domain.TransactionExpectation {
	o, ok := other.(transactionExpectation)
	if !ok {
		return te
	}
	combined := transactionExpectation{}
	for path, t := range te {
		combined[path] = t
	}
	for path, t := range o {
		existing, ok := combined[path]
		// if there already exists an expectation for the same document
		// choose the expecation with the later update time
		if !ok || existing.UpdateTime.Before(t.UpdateTime) {
			combined[path] = t
		}
	}
	return combined
}

func combinedId(boardId string, other string) string {
	return boardId + "." + other
}

func (f *firestoreBoardDataStore) UpdateBoard(ctx context.Context, boardId string, update *domain.DatastoreBoardUpdate) error {
	if update == nil || update.IsEmpty() {
		return fs.NewFirestoreError(nil, errors.InvalidArgument).WithInternalMessage("empty update, this might be a bug")
	}
	err := f.client.RunTransaction(ctx, func(c context.Context, t *firestore.Transaction) error {
		// verify transaction expectations
		te := update.TransactionExpecation
		if te != nil {
			lte, ok := te.(transactionExpectation)
			if !ok {
				return fs.NewFirestoreError(nil, errors.Internal).WithInternalMessage("passed transaction expectation of wrong type")
			}
			docRefs := fs.TransactionExpectations(lte).DocRefs()
			snaps, err := t.GetAll(docRefs)
			if err != nil {
				return fs.NewFirestoreError(nil, errors.Internal)
			}
			err = fs.TransactionExpectations(lte).Verify(snaps)
			if err != nil {
				return err
			}
		}

		// We only want to update a small number of users/invites in a map of users/invites.
		// And we also might want to remove a user/invite from the map of users/invites.
		// This is only possible by using a map, not with a fsBoardWithUsersAndInvites struct.
		u := map[string]interface{}{}
		users := map[string]interface{}{}
		invites := map[string]interface{}{}

		mergePaths := []firestore.FieldPath{}

		if update.UpdateBoard {
			u["board"] = newFsBoard(update.Board)
			mergePaths = append(mergePaths, firestore.FieldPath{"board"})
		}

		for _, user := range update.UpdateUsers {
			userId := user.User.UserId
			fsu := newFsBoardUser(user, boardId)
			users[userId] = fsu
			mergePaths = append(mergePaths, firestore.FieldPath{"users", userId})

			err := t.Set(f.boardUserCollection.Doc(combinedId(boardId, userId)), fsu)
			if err != nil {
				return err
			}
		}

		for _, userId := range update.RemoveUsers {
			users[userId] = firestore.Delete
			mergePaths = append(mergePaths, firestore.FieldPath{"users", userId})

			err := t.Delete(f.boardUserCollection.Doc(combinedId(boardId, userId)))
			if err != nil {
				return err
			}
		}

		for _, invite := range update.UpdateInvites {
			inviteId := invite.InviteId
			fsi := newFsBoardInvite(invite, boardId)
			invites[inviteId] = fsi
			mergePaths = append(mergePaths, firestore.FieldPath{"invites", inviteId})

			err := t.Set(f.boardInviteCollection.Doc(combinedId(boardId, inviteId)), fsi)
			if err != nil {
				return err
			}
		}

		for _, inviteId := range update.RemoveInvites {
			invites[inviteId] = firestore.Delete
			mergePaths = append(mergePaths, firestore.FieldPath{"invites", inviteId})

			err := t.Delete(f.boardInviteCollection.Doc(combinedId(boardId, inviteId)))
			if err != nil {
				return err
			}
		}

		if len(users) > 0 {
			u["users"] = users
		}
		if len(invites) > 0 {
			u["invites"] = invites
		}

		err := t.Set(f.boardCollection.Doc(boardId), u, firestore.Merge(mergePaths...))
		if err != nil {
			return err
		}

		return nil
	}, firestore.MaxAttempts(1))
	return err
}

func (f *firestoreBoardDataStore) DeleteBoard(ctx context.Context, boardId string) error {
	err := f.client.RunTransaction(ctx, func(c context.Context, t *firestore.Transaction) error {
		var b fsBoardWithUsersAndInvites
		snap, err := t.Get(f.boardCollection.Doc(boardId))
		if err != nil {
			return err
		}
		err = fs.UnmarshalDocSnapshot(snap, &b)
		if err != nil {
			return err
		}

		for userId := range b.Users {
			err = t.Delete(f.boardUserCollection.Doc(combinedId(boardId, userId)))
			if err != nil {
				return err
			}
		}

		for inviteId := range b.Invites {
			err = t.Delete(f.boardInviteCollection.Doc(combinedId(boardId, inviteId)))
			if err != nil {
				return err
			}
		}

		err = t.Delete(f.boardCollection.Doc(boardId))
		if err != nil {
			return err
		}

		return nil
	}, firestore.MaxAttempts(1))
	return err
}

func (f *firestoreBoardDataStore) Board(ctx context.Context, boardId string) (domain.BoardWithUsersAndInvites, domain.TransactionExpectation, error) {
	var board fsBoardWithUsersAndInvites
	te, err := fs.GetDocumentByIdWithTE(ctx, f.boardCollection, boardId, &board)
	if err != nil {
		return domain.BoardWithUsersAndInvites{}, nil, err
	}

	return newDomainBoardWithUsersAndInvites(board), transactionExpectation(te), nil
}

func (f *firestoreBoardDataStore) Boards(ctx context.Context, boardIds []string) ([]domain.Board, error) {
	if len(boardIds) == 0 {

	}

	docRefs := make([]*firestore.DocumentRef, len(boardIds))
	for i, boardId := range boardIds {
		docRefs[i] = f.boardCollection.Doc(boardId)
	}

	snaps, err := f.client.GetAll(ctx, docRefs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Board, 0, len(snaps))
	for _, snap := range snaps {
		if snap.Exists() {
			var board fsBoardWithUsersAndInvites
			//TODO can this be optimized? we dont really need to unmarshal the users and invites, we just need the board
			err := fs.UnmarshalDocSnapshot(snap, &board)
			if err != nil {
				return nil, err
			}
			result = append(result, newDomainBoard(board.Board))
		}
	}

	return result, nil
}

func (f *firestoreBoardDataStore) BoardsForUser(ctx context.Context, userId string, qp domain.QueryParams) ([]domain.Board, error) {
	query := f.boardUserCollection.Where("user.userId", "==", userId).Select("boardId").OrderBy("createdTime", firestore.Desc)
	if qp.Cursor != 0 {
		query = query.StartAt(qp.Cursor)
	}
	if qp.Limit != 0 {
		query = query.Limit(qp.Limit)
	}
	snaps, err := fs.GetDocumentsForQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(snaps) == 0 {
		return []domain.Board{}, nil
	}

	boardIds := make([]string, len(snaps))
	for i, snap := range snaps {
		v, err := snap.DataAt("boardId")
		if err != nil {
			return nil, err
		}
		boardId, ok := v.(string)
		if ok {
			boardIds[i] = boardId
		}
	}

	return f.Boards(ctx, boardIds)
}

func (f *firestoreBoardDataStore) User(ctx context.Context, boardId string, userId string) (domain.BoardUser, error) {
	var boardUser fsBoardUser
	err := fs.GetDocumentById(ctx, f.boardUserCollection, combinedId(boardId, userId), &boardUser)
	if err != nil {
		return domain.BoardUser{}, err
	}
	return newDomainBoardUser(boardUser), nil
}

func (f *firestoreBoardDataStore) InvitesForUser(ctx context.Context, userId string, qp domain.QueryParams) (map[string]domain.BoardInvite, error) {
	query := f.boardInviteCollection.Where("user.userId", "==", userId).OrderBy("createdTime", firestore.Desc)
	if qp.Cursor != 0 {
		query = query.StartAt(qp.Cursor)
	}
	if qp.Limit != 0 {
		query = query.Limit(qp.Limit)
	}
	snaps, err := fs.GetDocumentsForQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make(map[string]domain.BoardInvite, len(snaps))
	for _, snap := range snaps {
		var invite fsBoardInvite
		err = fs.UnmarshalDocSnapshot(snap, &invite)
		if err != nil {
			return nil, err
		}
		result[invite.BoardId] = newDomainBoardInvite(invite)
	}

	return result, nil
}
