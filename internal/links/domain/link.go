package domain

import (
	"context"
	"net/url"

	"github.com/d39b/kit/errors"
	"github.com/d39b/kit/time"
	"github.com/d39b/kit/uuid"
)

const (
	errTitleEmpty = iota + 1
	errTitleTooLong
	errUrlEmpty
	errUrlInvalid
	errInvalidRating
)

type User struct {
	UserId string
	Name   string
}

type Link struct {
	BoardId string
	LinkId  string

	Title string
	Url   string

	CreatedTime int64
	CreatedBy   User
}

func newError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "links/domain", code)
}

const maxTitleLength = 200

func (l *Link) IsValid() error {
	if len(l.Title) == 0 {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("title cannot be empty").WithPublicCode(errTitleEmpty)
	}

	if len(l.Title) > maxTitleLength {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("title too long").WithPublicCode(errTitleTooLong)
	}

	if len(l.Url) == 0 {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("url cannot be empty").WithPublicCode(errUrlEmpty)
	}

	u, err := url.Parse(l.Url)
	if err != nil {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("invalid url").WithPublicCode(errUrlInvalid)
	}
	// Only allow secure links
	// In an actual application we would probably perform other validations as well, e.g.
	// filtering out certain hosts (e.g. spam) ...
	if u.Scheme != "https" {
		return newError(nil, errors.InvalidArgument).WithPublicMessage("invalid url").WithPublicCode(errUrlInvalid)
	}
	return nil
}

func NewLink(boardId, title, url string, user User) (Link, error) {
	id, err := uuid.NewUUIDWithPrefix("l")
	if err != nil {
		return Link{}, newError(err, errors.Internal).WithInternalMessage("could not create link uuid")
	}

	timeNow := time.CurrTimeUnixNano()

	link := Link{
		BoardId:     boardId,
		LinkId:      id,
		Title:       title,
		Url:         url,
		CreatedTime: timeNow,
		CreatedBy:   user,
	}

	err = link.IsValid()
	if err != nil {
		return Link{}, err
	}

	return link, nil
}

// A summary/aggregate of all the ratings for a link
type Rating struct {
	// Score = Upvotes + Downvotes
	Score int
	// Total of positive ratings, must be >= 0
	Upvotes int
	// Total of negative ratings, must be <= 0
	Downvotes int
}

// The rating of a user for a link.
type UserLinkRating struct {
	UserId string
	// +1 for upvote, -1 for downvote
	Rating       int
	ModifiedTime int64
}

func NewUserLinkRating(rating int, userId string) (UserLinkRating, error) {
	if !(rating == 1 || rating == -1) {
		return UserLinkRating{}, newError(nil, errors.InvalidArgument).WithPublicMessage("invalid rating").WithPublicCode(errInvalidRating)
	}

	timeNow := time.CurrTimeUnixNano()

	return UserLinkRating{
		UserId:       userId,
		Rating:       rating,
		ModifiedTime: timeNow,
	}, nil
}

type LinkService struct {
	ds LinkDataStore
}

func NewLinkService(ds LinkDataStore) *LinkService {
	return &LinkService{ds: ds}
}

func newServiceError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "LinkService", code)
}

func (ls *LinkService) CreateLink(ctx context.Context, boardId string, title string, url string, user User) (Link, error) {
	link, err := NewLink(boardId, title, url, user)
	if err != nil {
		return Link{}, err
	}

	err = ls.ds.CreateLink(ctx, boardId, link)
	if err != nil {
		return Link{}, newServiceError(err, errors.Internal).WithInternalMessage("could not create link")
	}

	return link, nil
}

// Creates or changes the rating of a user for a link.
func (ls *LinkService) UpdateUserRating(ctx context.Context, boardId string, linkId string, rating int, user User) (UserLinkRating, error) {
	r, err := NewUserLinkRating(rating, user.UserId)
	if err != nil {
		return UserLinkRating{}, err
	}

	err = ls.ds.UpdateRating(ctx, boardId, linkId, r)
	if err != nil {
		return UserLinkRating{}, newServiceError(err, errors.Internal).WithInternalMessage("could not update link rating")
	}

	return r, nil
}
