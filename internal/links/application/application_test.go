package application

import (
	"context"
	"linkboards/internal/auth"
	"linkboards/internal/links/datastore/inmem"
	"linkboards/internal/links/domain"
	"testing"
	stdtime "time"

	"github.com/d39b/kit/errors"
	"github.com/d39b/kit/time"

	"github.com/stretchr/testify/assert"
)

type testUser struct {
	UserId string
	Name   string
	Role   string
}

var testUser1 = testUser{
	UserId: "user-1",
	Name:   "User One",
	Role:   auth.BoardRoleViewer,
}

var testUser2 = testUser{
	UserId: "user-2",
	Name:   "User Two",
	Role:   auth.BoardRoleEditor,
}

var testUsers = []testUser{testUser1, testUser2}

var testRole = "testRole"

type testAuthorizationStore struct{}

func (t *testAuthorizationStore) Roles(ctx context.Context, boardId string, userId string) ([]string, error) {
	for _, user := range testUsers {
		if user.UserId == userId {
			return []string{user.Role, testRole}, nil
		}
	}
	return nil, errors.New(nil, "test", errors.NotFound)
}

func newTestLinkDatastore() domain.LinkDataStore {
	return inmem.NewInmemLinkDataStore()
}

func TestUnauthenticatedUsersDenied(t *testing.T) {
	a := assert.New(t)

	// Uauthenticated users are denied
	ctx := context.Background()
	service := NewLinkApplicationService(newTestLinkDatastore(), &testAuthorizationStore{})

	_, err := service.CreateLink(ctx, "b-123", NewLink{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	err = service.DeleteLink(ctx, "b-123", "l-123")
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	err = service.RateLink(ctx, "b-123", "l-123", LinkRating{Rating: 1})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.Link(ctx, "b-123", "l-123")
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))

	_, err = service.Links(ctx, "b-123", LinkQueryParams{})
	a.NotNil(err)
	a.True(errors.IsUnauthenticatedError(err))
}

func TestAuthorization(t *testing.T) {
	a := assert.New(t)

	ds := newTestLinkDatastore()

	// This will create a new service that uses an AuthorizationChecker
	// that gives the user all scopes except the given one
	newService := func(withoutScope auth.Scope) LinkApplicationService {
		as := allScopes()
		scopes := make([]auth.Scope, 0, len(as)-1)
		for _, scope := range as {
			if scope != withoutScope {
				scopes = append(scopes, scope)
			}
		}

		rts := map[string][]auth.Scope{
			testRole: scopes,
		}

		service := &linkApplicationService{
			linkService:   domain.NewLinkService(ds),
			linkDataStore: ds,
			authChecker:   auth.NewAuthorizationChecker(rts, &testAuthorizationStore{}),
		}
		return service
	}

	ctx := auth.ContextWithUser(context.Background(), auth.User{
		UserId: testUser1.UserId,
		Name:   testUser1.Name,
	})

	err := ds.CreateLink(context.Background(), "b-123", domain.Link{
		LinkId: "l-123",
		CreatedBy: domain.User{
			UserId: testUser2.UserId,
		},
	})
	a.Nil(err)

	_, err = newService(createLinkScope).CreateLink(ctx, "b-123", NewLink{})
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	err = newService(deleteLinkScope).DeleteLink(ctx, "b-123", "l-123")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	err = newService(rateLinkScope).RateLink(ctx, "b-123", "l-123", LinkRating{Rating: 1})
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	_, err = newService(queryLinksScope).Link(ctx, "b-123", "l-123")
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))

	_, err = newService(queryLinksScope).Links(ctx, "b-123", LinkQueryParams{})
	a.NotNil(err)
	a.True(errors.IsPermissionDeniedError(err))
}

var defaultTime = stdtime.Date(2022, 04, 04, 0, 0, 0, 0, stdtime.UTC)
var defaultTimeUnix = defaultTime.UnixNano()

func TestLinks(t *testing.T) {
	a := assert.New(t)

	time.TimeFunc = func() stdtime.Time {
		return defaultTime
	}

	ctx := auth.ContextWithUser(context.Background(), auth.User{
		UserId: testUser1.UserId,
		Name:   testUser1.Name,
	})
	svc := NewLinkApplicationService(newTestLinkDatastore(), &testAuthorizationStore{})

	link, err := svc.CreateLink(ctx, "b-123", NewLink{Title: "Link title", Url: "https://abc.com/xyz"})
	a.Nil(err)
	a.NotEmpty(link.LinkId)
	a.Equal("b-123", link.BoardId)
	a.Equal("Link title", link.Title)
	a.Equal("https://abc.com/xyz", link.Url)
	a.Equal(0, link.Score)
	a.Equal(0, link.Downvotes)
	a.Equal(0, link.Upvotes)
	a.Equal(0, link.UserRating)
	a.Equal(defaultTimeUnix, link.CreatedTime)
	a.Equal(testUser1.UserId, link.CreatedBy.UserId)

	// rate the link
	err = svc.RateLink(ctx, "b-123", link.LinkId, LinkRating{Rating: 1})
	a.Nil(err)

	// get link again
	ratedLink, err := svc.Link(ctx, "b-123", link.LinkId)
	a.Nil(err)
	a.NotEmpty(ratedLink.LinkId)
	a.Equal("b-123", ratedLink.BoardId)
	a.Equal("Link title", ratedLink.Title)
	a.Equal("https://abc.com/xyz", ratedLink.Url)
	a.Equal(1, ratedLink.Score)
	a.Equal(0, ratedLink.Downvotes)
	a.Equal(1, ratedLink.Upvotes)
	a.Equal(1, ratedLink.UserRating)

	// query links
	links, err := svc.Links(ctx, "b-123", LinkQueryParams{})
	a.Nil(err)
	a.NotEmpty(links)
	ql := links[0]
	a.Equal(link.LinkId, ql.LinkId)
	a.Equal(1, ql.Score)
	a.Equal(1, ql.UserRating)

	// delete it
	err = svc.DeleteLink(ctx, "b-123", link.LinkId)
	a.Nil(err)

	err = svc.DeleteLink(ctx, "b-123", link.LinkId)
	a.NotNil(err)
	a.True(errors.IsNotFoundError(err))
}

func TestLinkQueryParam(t *testing.T) {
	a := assert.New(t)

	score := 10
	createdTime := int64(133742)
	lqp := LinkQueryParams{
		Limit:             50,
		Sort:              "top",
		CursorScore:       &score,
		CursorCreatedTime: &createdTime,
	}

	dlp := lqp.toDatastoreQueryParams()
	a.Equal(50, dlp.Limit)
	a.EqualValues(domain.SortOrderTop, dlp.SortOrder)
	a.Equal(10, *dlp.CursorScore)
	a.EqualValues(133742, dlp.CursorCreatedTime)
}
