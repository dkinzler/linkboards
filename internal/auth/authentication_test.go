package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanAddAndRetrieveUserFromContext(t *testing.T) {
	a := assert.New(t)
	user := User{UserId: "u-1", Name: "User One"}

	ctx := ContextWithUser(context.Background(), user)

	u, ok := UserFromContext(ctx)
	a.True(ok)
	a.Equal(user, u)
}
