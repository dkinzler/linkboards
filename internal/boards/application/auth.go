package application

import (
	"context"

	"github.com/d39b/linkboards/internal/auth"
	"github.com/d39b/linkboards/internal/boards/domain"
)

func userFromContext(ctx context.Context) (auth.User, bool) {
	return auth.UserFromContext(ctx)
}

func toDomainUser(u auth.User) domain.User {
	return domain.User{
		UserId: u.UserId,
		Name:   u.Name,
	}
}

const (
	createBoardScope         auth.Scope = "boards:create"
	deleteBoardScope                    = "boards:delete"
	editBoardScope                      = "boards:edit"
	viewBoardScope                      = "boards:view"
	viewBoardUsersScope                 = "boards:viewUsers"
	viewBoardInvitesScope               = "boards:viewInvites"
	listUserBoardsScope                 = "boards:list"
	createInviteScope                   = "boards:invite"
	deleteInviteScope                   = "boards:deleteInvite"
	respondToInviteScope                = "boards:respondToInvite"
	listUserInvitesScope                = "boards:listUserInvites"
	removeUserFromBoardScope            = "boards:removeUser"
	editBoardUserScope                  = "boards:editUsers"
)

func allScopes() []auth.Scope {
	return []auth.Scope{
		createBoardScope,
		deleteBoardScope,
		editBoardScope,
		viewBoardScope,
		viewBoardUsersScope,
		viewBoardInvitesScope,
		listUserBoardsScope,
		createInviteScope,
		deleteInviteScope,
		respondToInviteScope,
		listUserInvitesScope,
		removeUserFromBoardScope,
		editBoardUserScope,
	}
}

var roleToScopes = map[string][]auth.Scope{
	auth.BoardRoleOwner: {
		deleteBoardScope,
		editBoardScope,
		viewBoardScope,
		viewBoardUsersScope,
		viewBoardInvitesScope,
		removeUserFromBoardScope,
		editBoardUserScope,
		createInviteScope,
		deleteInviteScope,
		respondToInviteScope,
	},
	auth.BoardRoleEditor: {
		editBoardScope,
		viewBoardScope,
		viewBoardUsersScope,
		viewBoardInvitesScope,
		removeUserFromBoardScope,
		editBoardUserScope,
		createInviteScope,
		deleteInviteScope,
		respondToInviteScope,
	},
	auth.BoardRoleViewer: {
		viewBoardScope,
		respondToInviteScope,
	},
}

// scopes available to authenticated users
var authenticatedScopes auth.Authorization = map[auth.Scope]bool{
	createBoardScope:     true,
	listUserBoardsScope:  true,
	respondToInviteScope: true,
	listUserInvitesScope: true,
}

func NewAuthorizationChecker(as auth.AuthorizationStore) *auth.BoardAuthorizationChecker {
	return auth.NewAuthorizationChecker(roleToScopes, as)
}
