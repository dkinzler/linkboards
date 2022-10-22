package auth

import (
	"context"
)

// Authenticated user making a request to a component/application service.
type User struct {
	UserId string
	Name   string
}

type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

// Used to add a user to context, from which it can be retrieved by the application service.
// E.g. a transport or endpoint middleware could decode a JWT token contained in an http request header and
// use it to create a User instance.
func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
