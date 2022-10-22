package middleware

import (
	"context"
	"encoding/base64"
	"go-sample/internal/auth"
	"strings"

	"github.com/d39b/kit/errors"
	lfbauth "github.com/d39b/kit/firebase/auth"

	fbauth "firebase.google.com/go/v4/auth"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http"
)

// Can be used for development/testing.
// Simply provide a non-empty user id and any password in the http authorization header
// using the basic authentication format (i.e. "Basic base64Encode(userId:password)").
// Make sure to use the "PopulateRequestContext" RequestFunc from the go-kit/kit/transport/http package as a "ServerBefore" option when creating http handlers.
func NewFakeAuthEndpointMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			authorizationHeader, ok := ctx.Value(http.ContextKeyRequestAuthorization).(string)
			if !ok {
				return nil, errors.New(nil, "fakeAuthMiddleware", errors.Unauthenticated).
					WithPublicMessage("empty Authorization header")
			}

			prefix := "Basic "
			if !strings.HasPrefix(authorizationHeader, prefix) {
				return nil, errors.New(nil, "fakeAuthMiddleware", errors.Unauthenticated).
					WithPublicMessage("invalid Authorization header")
			}

			s, err := base64.StdEncoding.DecodeString(authorizationHeader[len(prefix):])
			if err != nil {
				return nil, errors.New(err, "fakeAuthMiddleware", errors.Unauthenticated).
					WithPublicMessage("invalid Authorization header")
			}
			userId, _, ok := strings.Cut(string(s), ":")
			if !ok {
				return nil, errors.New(nil, "fakeAuthMiddleware", errors.Unauthenticated).WithInternalMessage("basic auth string does not contain :").
					WithPublicMessage("invalid Authorization header")
			}

			if userId == "" {
				return nil, errors.New(nil, "fakeAuthMiddleware", errors.Unauthenticated).WithInternalMessage("empty user id").
					WithPublicMessage("invalid Authorization header")
			}

			newCtx := auth.ContextWithUser(ctx, auth.User{
				UserId: userId,
			})

			return next(newCtx, request)
		}
	}
}

func NewFirebaseAuthEndpointMiddleware(authClient *fbauth.Client, requireVerifiedEmail bool) endpoint.Middleware {
	fbAuthChecker := lfbauth.NewAuthChecker(authClient, requireVerifiedEmail, nil)
	return lfbauth.NewAuthEndpointMiddleware(
		fbAuthChecker,
		func(ctx context.Context, u lfbauth.User) context.Context {
			return auth.ContextWithUser(ctx, auth.User{
				UserId: u.Uid,
			})
		},
	)
}
