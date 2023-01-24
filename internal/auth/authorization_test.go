package auth

import (
	"context"
	"testing"

	"github.com/dkinzler/kit/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsBoardRoleValid(t *testing.T) {
	a := assert.New(t)

	for _, role := range BoardRoles {
		a.True(IsBoardRoleValid(role))
	}

	a.False(IsBoardRoleValid("notavalidrole"))
}

type testUser struct {
	UserId string
	Roles  []string
}

var testUser1 = testUser{
	UserId: "u-123",
	Roles:  []string{BoardRoleOwner},
}

var testUser2 = testUser{
	UserId: "u-456",
	Roles:  []string{BoardRoleViewer, BoardRoleEditor},
}

var testUsers = []testUser{testUser1, testUser2}

type testAuthorizationStore struct{}

func (t *testAuthorizationStore) Roles(ctx context.Context, boardId string, userId string) ([]string, error) {
	for _, user := range testUsers {
		if user.UserId == userId {
			return user.Roles, nil
		}
	}
	return nil, errors.New(nil, "test", errors.NotFound)
}

func TestBoardAuthorizationChecker(t *testing.T) {
	a := assert.New(t)

	authStore := &testAuthorizationStore{}
	rolesToScope := map[string][]Scope{
		BoardRoleOwner:  {"ownerScope1", "ownerScope2"},
		BoardRoleEditor: {"editorScope1", "editorScope2"},
		BoardRoleViewer: {"viewerScope1", "viewerScope2"},
	}
	checker := NewAuthorizationChecker(rolesToScope, authStore)
	ctx := context.Background()

	// if user not found, should return error
	az, err := checker.GetAuthorization(ctx, "b-123", "usernotonboard")
	a.NotNil(err)
	a.Empty(az)

	// correct scopes are returned
	az, err = checker.GetAuthorization(ctx, "b-123", testUser1.UserId)
	a.Nil(err)
	a.True(az.HasScope("ownerScope1"))
	a.True(az.HasScope("ownerScope2"))
	a.False(az.HasScope("editorScope2"))

	az, err = checker.GetAuthorization(ctx, "b-123", testUser2.UserId)
	a.Nil(err)
	a.False(az.HasScope("ownerScope1"))
	a.False(az.HasScope("ownerScope2"))
	a.True(az.HasScope("editorScope1"))
	a.True(az.HasScope("editorScope2"))
	a.True(az.HasScope("viewerScope1"))
	a.True(az.HasScope("viewerScope2"))
}
