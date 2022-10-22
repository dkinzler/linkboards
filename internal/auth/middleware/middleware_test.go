package middleware

import (
	"context"
	"encoding/base64"
	"go-sample/internal/auth"
	"testing"

	"github.com/d39b/kit/errors"
	"github.com/go-kit/kit/transport/http"
	"github.com/stretchr/testify/assert"
)

func TestFakeAuthEndpointMiddleware(t *testing.T) {
	a := assert.New(t)

	mw := NewFakeAuthEndpointMiddleware()

	called := false
	e := func(ctx context.Context, request interface{}) (interface{}, error) {
		called = true
		user, ok := auth.UserFromContext(ctx)
		if ok {
			return user, nil
		}
		return nil, nil
	}

	_, err := mw(e)(context.Background(), nil)
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))
	a.False(called)

	ctx := context.WithValue(context.Background(), http.ContextKeyRequestAuthorization, "this isnt correct format")
	_, err = mw(e)(ctx, nil)
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))
	a.False(called)

	// correct header but empty userId shouldn't work
	v := base64.StdEncoding.EncodeToString([]byte(":somepw"))
	ctx = context.WithValue(context.Background(), http.ContextKeyRequestAuthorization, "Basic "+v)
	_, err = mw(e)(ctx, nil)
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))
	a.False(called)

	v = base64.StdEncoding.EncodeToString([]byte("u-123-456"))
	ctx = context.WithValue(context.Background(), http.ContextKeyRequestAuthorization, "Basic "+v)
	_, err = mw(e)(ctx, nil)
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))
	a.False(called)

	// with a correct header it should work
	v = base64.StdEncoding.EncodeToString([]byte("u-123-456:somepw"))
	ctx = context.WithValue(context.Background(), http.ContextKeyRequestAuthorization, "Basic "+v)
	user, err := mw(e)(ctx, nil)
	a.Nil(err)
	a.True(called)
	authUser, ok := user.(auth.User)
	a.True(ok)
	a.Equal("u-123-456", authUser.UserId)
}
