package domain

import (
	"context"
	"testing"

	"github.com/d39b/kit/errors"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

type MockLinkDataStore struct {
	mock.Mock
}

func (m *MockLinkDataStore) CreateLink(ctx context.Context, boardId string, link Link) error {
	args := m.Called(boardId, link)
	return args.Error(0)
}

func (m *MockLinkDataStore) DeleteLink(ctx context.Context, boardId string, linkId string) error {
	args := m.Called(boardId, linkId)
	return args.Error(0)
}

func (m *MockLinkDataStore) UpdateRating(ctx context.Context, boardId string, linkId string, rating UserLinkRating) error {
	args := m.Called(boardId, linkId, rating)
	return args.Error(0)
}

func (m *MockLinkDataStore) Link(ctx context.Context, boardId string, linkId string, rf LinkReturnFields) (LinkWithRating, error) {
	args := m.Called(boardId, linkId, rf)
	return args.Get(0).(LinkWithRating), args.Error(1)
}

func (m *MockLinkDataStore) Links(ctx context.Context, boardId string, rf LinkReturnFields, qp LinkQueryParams) ([]LinkWithRating, error) {
	args := m.Called(boardId, rf, qp)
	return args.Get(0).([]LinkWithRating), args.Error(1)
}

func TestLinkCreation(t *testing.T) {
	a := assert.New(t)

	ds := &MockLinkDataStore{}
	svc := NewLinkService(ds)
	ctx := context.Background()

	// empty title shouldn't work
	link, err := svc.CreateLink(ctx, "b-123", "", "https://example.com/awesome.png", User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errTitleEmpty))
	a.Empty(link)

	// empty link shouldnt work
	link, err = svc.CreateLink(ctx, "b-123", "A title", "", User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errUrlEmpty))
	a.Empty(link)

	// invalid url
	link, err = svc.CreateLink(ctx, "b-123", "A title", "$%&/\\:\\12$4%", User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errUrlInvalid))
	a.Empty(link)

	// invalid url, no https
	link, err = svc.CreateLink(ctx, "b-123", "A title", "http://example.com", User{UserId: "u-123"})
	a.NotNil(err)
	a.True(errors.HasPublicCode(err, errUrlInsecure))
	a.Empty(link)

	// this should work
	ds.On("CreateLink", "b-123", mock.MatchedBy(func(l Link) bool {
		if l.Title != "A title" {
			return false
		}
		if l.Url != "https://example.com" {
			return false
		}
		return true
	})).Return(nil).Once()
	link, err = svc.CreateLink(ctx, "b-123", "A title", "https://example.com", User{UserId: "u-123"})
	a.Nil(err)
	a.Equal("A title", link.Title)
	a.Equal("https://example.com", link.Url)
	a.Equal("b-123", link.BoardId)
	a.NotEmpty(link.LinkId)
	a.NotZero(link.CreatedTime)
	a.Equal(User{UserId: "u-123"}, link.CreatedBy)
	ds.AssertExpectations(t)

	// fails on datastore error
	ds.On("CreateLink", "b-123", mock.Anything).Return(errors.New(nil, "test", errors.Internal)).Once()
	link, err = svc.CreateLink(ctx, "b-123", "A title", "https://example.com", User{UserId: "u-123"})
	a.Empty(link)
	a.NotNil(err)
	ds.AssertExpectations(t)
}

func TestRating(t *testing.T) {
	a := assert.New(t)

	ds := &MockLinkDataStore{}
	svc := NewLinkService(ds)
	ctx := context.Background()

	// invalid rating
	rating, err := svc.UpdateUserRating(ctx, "b-123", "l-123", 0, User{UserId: "u-123"})
	a.NotNil(err)
	a.Empty(rating)
	a.True(errors.HasPublicCode(err, errInvalidRating))

	// fails if data store call fails
	ds.On("UpdateRating", "b-123", "l-123", mock.Anything).Return(errors.New(nil, "test", errors.Internal)).Once()
	rating, err = svc.UpdateUserRating(ctx, "b-123", "l-123", 1, User{UserId: "u-123"})
	a.NotNil(err)
	a.Empty(rating)

	// should work
	ds.On("UpdateRating", "b-123", "l-123", mock.MatchedBy(func(r UserLinkRating) bool {
		if r.Rating != 1 {
			return false
		}
		if r.UserId != "u-123" {
			return false
		}
		return true
	})).Return(nil).Once()
	rating, err = svc.UpdateUserRating(ctx, "b-123", "l-123", 1, User{UserId: "u-123"})
	a.Nil(err)
	a.Equal("u-123", rating.UserId)
	a.Equal(1, rating.Rating)
	a.NotZero(rating.ModifiedTime)
}
