package firestore

import "github.com/d39b/linkboards/internal/boards/domain"

// We define separate types here, since the way we store the data in firestore
// is slightly different from the domain types.
// This approach also avoids firestore struct tags on domain types,
// which would leak implementation details from this package into the domain package.

type fsUser struct {
	UserId string `firestore:"userId"`
	Name   string `firestore:"name"`
}

func newFsUser(u domain.User) fsUser {
	return fsUser{
		UserId: u.UserId,
		Name:   u.Name,
	}
}

type fsBoard struct {
	BoardId      string `firestore:"boardId"`
	Name         string `firestore:"name"`
	Description  string `firestore:"description"`
	CreatedTime  int64  `firestore:"createdTime"`
	CreatedBy    fsUser `firestore:"createdBy"`
	ModifiedTime int64  `firestore:"modifiedTime"`
	ModifiedBy   fsUser `firestore:"modifiedBy"`
}

func newFsBoard(board domain.Board) fsBoard {
	return fsBoard{
		BoardId:      board.BoardId,
		Name:         board.Name,
		Description:  board.Description,
		CreatedTime:  board.CreatedTime,
		CreatedBy:    newFsUser(board.CreatedBy),
		ModifiedTime: board.ModifiedTime,
		ModifiedBy:   newFsUser(board.ModifiedBy),
	}
}

type fsBoardUser struct {
	BoardId      string `firestore:"boardId"`
	User         fsUser `firestore:"user"`
	Role         string `firestore:"role"`
	CreatedTime  int64  `firestore:"createdTime"`
	InvitedBy    fsUser `firestore:"invitedBy"`
	ModifiedTime int64  `firestore:"modifiedTime"`
	ModifiedBy   fsUser `firestore:"modifiedBy"`
}

func newFsBoardUser(bu domain.BoardUser, boardId string) fsBoardUser {
	return fsBoardUser{
		BoardId:      boardId,
		User:         newFsUser(bu.User),
		Role:         bu.Role,
		CreatedTime:  bu.CreatedTime,
		InvitedBy:    newFsUser(bu.InvitedBy),
		ModifiedTime: bu.ModifiedTime,
		ModifiedBy:   newFsUser(bu.ModifiedBy),
	}
}

type fsBoardInvite struct {
	BoardId     string `firestore:"boardId"`
	InviteId    string `firestore:"inviteId"`
	Role        string `firestore:"role"`
	User        fsUser `firestore:"user"`
	CreatedTime int64  `firestore:"createdTime"`
	CreatedBy   fsUser `firestore:"createdBy"`
	ExpiresTime int64  `firestore:"expiresTime"`
}

func newFsBoardInvite(i domain.BoardInvite, boardId string) fsBoardInvite {
	return fsBoardInvite{
		BoardId:     boardId,
		InviteId:    i.InviteId,
		Role:        i.Role,
		User:        newFsUser(i.User),
		CreatedTime: i.CreatedTime,
		CreatedBy:   newFsUser(i.CreatedBy),
		ExpiresTime: i.ExpiresTime,
	}
}

type fsBoardWithUsersAndInvites struct {
	Board fsBoard `firestore:"board"`

	Users   map[string]fsBoardUser   `firestore:"users"`
	Invites map[string]fsBoardInvite `firestore:"invites"`
}

func newDomainUser(fs fsUser) domain.User {
	return domain.User{
		UserId: fs.UserId,
		Name:   fs.Name,
	}
}

func newDomainBoard(fs fsBoard) domain.Board {
	return domain.Board{
		BoardId:      fs.BoardId,
		Name:         fs.Name,
		Description:  fs.Description,
		CreatedTime:  fs.CreatedTime,
		CreatedBy:    newDomainUser(fs.CreatedBy),
		ModifiedTime: fs.ModifiedTime,
		ModifiedBy:   newDomainUser(fs.ModifiedBy),
	}
}

func newDomainBoardUser(fs fsBoardUser) domain.BoardUser {
	return domain.BoardUser{
		User:         newDomainUser(fs.User),
		Role:         fs.Role,
		CreatedTime:  fs.CreatedTime,
		InvitedBy:    newDomainUser(fs.InvitedBy),
		ModifiedTime: fs.ModifiedTime,
		ModifiedBy:   newDomainUser(fs.ModifiedBy),
	}
}

func newDomainBoardInvite(fs fsBoardInvite) domain.BoardInvite {
	return domain.BoardInvite{
		InviteId:    fs.InviteId,
		Role:        fs.Role,
		User:        newDomainUser(fs.User),
		CreatedTime: fs.CreatedTime,
		CreatedBy:   newDomainUser(fs.CreatedBy),
		ExpiresTime: fs.ExpiresTime,
	}
}

func newDomainBoardWithUsersAndInvites(f fsBoardWithUsersAndInvites) domain.BoardWithUsersAndInvites {
	result := domain.BoardWithUsersAndInvites{
		Board: newDomainBoard(f.Board),
	}

	users := make([]domain.BoardUser, 0, len(f.Users))
	for _, user := range f.Users {
		users = append(users, newDomainBoardUser(user))
	}

	invites := make([]domain.BoardInvite, 0, len(f.Invites))
	for _, invite := range f.Invites {
		invites = append(invites, newDomainBoardInvite(invite))
	}

	result.Users = users
	result.Invites = invites
	return result
}
