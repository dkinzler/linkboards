package application

import (
	"github.com/dkinzler/linkboards/internal/auth"
	"github.com/dkinzler/linkboards/internal/links/domain"
)

func toDomainUser(u auth.User) domain.User {
	return domain.User{
		UserId: u.UserId,
		Name:   u.Name,
	}
}

const (
	createLinkScope auth.Scope = "links:create"
	deleteLinkScope            = "links:delete"
	rateLinkScope              = "links:rate"
	queryLinksScope            = "links:query"
)

func allScopes() []auth.Scope {
	return []auth.Scope{
		createLinkScope,
		deleteLinkScope,
		rateLinkScope,
		queryLinksScope,
	}
}

var roleToScopes = map[string][]auth.Scope{
	auth.BoardRoleOwner: {
		createLinkScope,
		deleteLinkScope,
		rateLinkScope,
		queryLinksScope,
	},
	auth.BoardRoleEditor: {
		createLinkScope,
		deleteLinkScope,
		rateLinkScope,
		queryLinksScope,
	},
	auth.BoardRoleViewer: {
		createLinkScope,
		rateLinkScope,
		queryLinksScope,
	},
}

// scopes available to authenticated users that are not a member of a board
var authenticatedScopes auth.Authorization = map[auth.Scope]struct{}{}

func NewAuthorizationChecker(as auth.AuthorizationStore) *auth.BoardAuthorizationChecker {
	return auth.NewAuthorizationChecker(roleToScopes, as)
}
