package auth

import (
	"context"
)

// Scopes are the most fine-grained units of permission and are defined
// directly in the components/bounded contexts that use them.
// They are usually named by combining the component and action/permission name, e.g. "comments:create" or "comments:delete".
type Scope string

// A role comprises a set of scopes.
const (
	// Roles a user can have for a board.
	// These roles are shared across features/components/bounded contexts of this app, since most of them require authorization relative to boards.
	// E.g. whether or not a user is authorized to invite other users, add/delete links or
	// post/delete comments to a board depends on the roles the user has for a particular board.
	BoardRoleOwner  = "owner"
	BoardRoleEditor = "editor"
	BoardRoleViewer = "viewer"
)

var BoardRoles = []string{BoardRoleOwner, BoardRoleEditor, BoardRoleViewer}

// Returns true if the given string denotes a valid board role.
func IsBoardRoleValid(role string) bool {
	for _, r := range BoardRoles {
		if r == role {
			return true
		}
	}
	return false
}

// Authorization represents the set of scopes a user has access to.
type Authorization map[Scope]bool

func (a Authorization) HasScope(scope Scope) bool {
	_, ok := a[scope]
	return ok
}

// AuthorizationStore can be used to get the roles a user has for a board.
type AuthorizationStore interface {
	Roles(ctx context.Context, boardId string, userId string) ([]string, error)
}

// BoardAuthorizationChecker can be used to obtain the set of scopes a user has access to for a given board.
type BoardAuthorizationChecker struct {
	roleToScopes map[string][]Scope
	store        AuthorizationStore
}

func NewAuthorizationChecker(roleToScopes map[string][]Scope, store AuthorizationStore) *BoardAuthorizationChecker {
	return &BoardAuthorizationChecker{
		roleToScopes: roleToScopes,
		store:        store,
	}
}

func (ac *BoardAuthorizationChecker) GetAuthorization(ctx context.Context, boardId string, userId string) (Authorization, error) {
	roles, err := ac.store.Roles(ctx, boardId, userId)
	if err != nil {
		return Authorization{}, err
	}

	scopes := make(map[Scope]bool)
	for _, role := range roles {
		scopesForRole, ok := ac.roleToScopes[role]
		if ok {
			for _, scope := range scopesForRole {
				scopes[scope] = true
			}
		}
	}
	return scopes, nil
}
