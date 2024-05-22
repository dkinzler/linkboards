package domain

import "context"

// Any type implementing this interface can be used as the link and rating data store for this application.
//
// Note that we do not aggregate the user ratings in the domain layer. Instead we expect that the data store implementation is capable
// of computing a Rating value that aggregates all the user ratings.
// How this is done is up to the implementation. One could e.g. store the aggregate explicitly and update it every time a user rating changes
// or it could be computed from all the user ratings and cached.
type LinkDataStore interface {
	CreateLink(ctx context.Context, boardId string, link Link) error
	DeleteLink(ctx context.Context, boardId string, linkId string) error

	UpdateRating(ctx context.Context, boardId string, linkId string, rating UserLinkRating) error

	// Return a single link, possibly with a rating summary and the rating for a particular user, depending on the LinkReturnFields value.
	Link(ctx context.Context, boardId string, linkId string, rf LinkReturnFields) (LinkWithRating, error)
	Links(ctx context.Context, boardId string, rf LinkReturnFields, qp LinkQueryParams) ([]LinkWithRating, error)
}

// For convenience, LinkDataStore should be able to return a link with its rating (i.e. the summary/aggregate of all user ratings) and possibly
// the individual rating of a given user.
// A LinkReturnFields value can be used to configure what should be returned.
type LinkWithRating struct {
	Link Link
	// Rating of the link, might be empty
	Rating Rating
	// Rating of the link from a user, might be empty
	UserRating UserLinkRating
}

// Defines what data a LinkDataStore query result should contain
type LinkReturnFields struct {
	// If true include the ratings of the link
	IncludeRating bool
	// If not empty include the rating of the user with the given userId
	IncludeUserRatingFor string
}

type SortOrder string

const SortOrderNewest = "newest"

// sort by rating
const SortOrderTop = "top"

type LinkQueryParams struct {
	// Should be either newest or top, defaults to newest
	SortOrder SortOrder
	// Maximum number of links to return.
	// Valid values are between 1 and 100, defaults to 20.
	Limit int

	// Query cursors that can be used for pagination.
	// When sort order is "newest", just provide CursorCreatedTime.
	// When sort order is "top", provide both values,
	// since there can possibly be many links with the same score.
	//
	// Return only links with less than or equal to score.
	// Use a pointer here to distinguish the case where no value is set, since 0 would be a valid cursor value.
	CursorScore *int
	// Return only links that were created at or before the given time (Unix nanoseconds).
	CursorCreatedTime int64
}

func NewLinkQueryParams() LinkQueryParams {
	return LinkQueryParams{
		SortOrder: SortOrderNewest,
		Limit:     20,
	}
}

func (l LinkQueryParams) SortByNewest() LinkQueryParams {
	l.SortOrder = SortOrderNewest
	return l
}

func (l LinkQueryParams) SortByTop() LinkQueryParams {
	l.SortOrder = SortOrderTop
	return l
}

func (l LinkQueryParams) WithLimit(limit int) LinkQueryParams {
	if limit > 0 && limit < 100 {
		l.Limit = limit
	}
	return l
}

func (l LinkQueryParams) WithScoreCursor(score int) LinkQueryParams {
	l.CursorScore = &score
	return l
}

func (l LinkQueryParams) WithCreatedTimeCursor(t int64) LinkQueryParams {
	if t > 0 {
		l.CursorCreatedTime = t
	}
	return l
}
