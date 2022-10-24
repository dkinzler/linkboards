// Package inmem provides an in-memory implementation of boards/domain/BoardDataStore, that
// can be used for development/testing.
package inmem

import (
	"context"
	"sort"
	"sync"

	"github.com/d39b/linkboards/internal/links/domain"

	"github.com/d39b/kit/errors"
)

type linkWithRatings struct {
	boardId     string
	link        domain.Link
	rating      domain.Rating
	userRatings map[string]domain.UserLinkRating
}

type InmemLinkDataStore struct {
	m sync.RWMutex
	// This storage setup assumes that link ids are unique across boards, which they are right now.
	links map[string]*linkWithRatings
}

func NewInmemLinkDataStore() *InmemLinkDataStore {
	return &InmemLinkDataStore{
		links: make(map[string]*linkWithRatings),
	}
}

func newError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "InmemLinkDataStore", code)
}

func (ds *InmemLinkDataStore) CreateLink(ctx context.Context, boardId string, link domain.Link) error {
	ds.m.Lock()
	defer ds.m.Unlock()

	l := &linkWithRatings{}
	l.boardId = boardId
	l.link = link
	l.rating = domain.Rating{}
	l.userRatings = make(map[string]domain.UserLinkRating)

	ds.links[link.LinkId] = l

	return nil
}

func (ds *InmemLinkDataStore) DeleteLink(ctx context.Context, boardId string, linkId string) error {
	ds.m.Lock()
	defer ds.m.Unlock()

	delete(ds.links, linkId)
	return nil
}

func (ds *InmemLinkDataStore) UpdateRating(ctx context.Context, boardId string, linkId string, rating domain.UserLinkRating) error {
	ds.m.Lock()
	defer ds.m.Unlock()

	l, ok := ds.links[linkId]
	if !ok {
		return newError(nil, errors.NotFound)
	}

	aggregate := l.rating

	oldRating, ok := l.userRatings[rating.UserId]
	if ok {
		if oldRating.Rating > 0 {
			aggregate.Upvotes -= oldRating.Rating
		} else {
			aggregate.Downvotes += oldRating.Rating
		}
	}

	if rating.Rating > 0 {
		aggregate.Upvotes += rating.Rating
	} else {
		aggregate.Downvotes += rating.Rating
	}

	aggregate.Score = aggregate.Upvotes + aggregate.Downvotes
	l.rating = aggregate
	l.userRatings[rating.UserId] = rating

	return nil
}

func (ds *InmemLinkDataStore) Link(ctx context.Context, boardId string, linkId string, rf domain.LinkReturnFields) (domain.LinkWithRating, error) {
	l, ok := ds.links[linkId]
	if !ok {
		return domain.LinkWithRating{}, newError(nil, errors.NotFound)
	}

	return linkWithRatingFromReturnFields(l, rf), nil
}

func (ds *InmemLinkDataStore) Links(ctx context.Context, boardId string, rf domain.LinkReturnFields, qp domain.LinkQueryParams) ([]domain.LinkWithRating, error) {
	links := make([]*linkWithRatings, 0)

	for _, link := range ds.links {
		if link.boardId == boardId {
			match := true

			if qp.CursorCreatedTime != 0 {
				if link.link.CreatedTime > qp.CursorCreatedTime {
					match = false
				}
			}

			if qp.CursorScore != nil {
				if link.rating.Score > *qp.CursorScore {
					match = false
				}
			}

			if match {
				links = append(links, link)
			}
		}
	}

	if qp.SortOrder == domain.SortOrderNewest {
		sort.Slice(links, func(i int, j int) bool {
			if links[i].link.CreatedTime > links[j].link.CreatedTime {
				return true
			}
			return false
		})
	} else if qp.SortOrder == domain.SortOrderTop {
		sort.Slice(links, func(i int, j int) bool {
			if links[i].rating.Score > links[j].rating.Score {
				return true
			} else if links[i].rating.Score == links[j].rating.Score {
				return links[i].link.CreatedTime > links[j].link.CreatedTime
			}
			return false
		})
	}

	if qp.Limit > 0 && qp.Limit < len(links) {
		links = links[0:qp.Limit]
	}

	result := make([]domain.LinkWithRating, len(links))
	for i, link := range links {
		result[i] = linkWithRatingFromReturnFields(link, rf)
	}

	return result, nil
}

func linkWithRatingFromReturnFields(l *linkWithRatings, rf domain.LinkReturnFields) domain.LinkWithRating {
	result := domain.LinkWithRating{
		Link: l.link,
	}

	if rf.IncludeRating {
		result.Rating = l.rating
	}

	if rf.IncludeUserRatingFor != "" {
		userRating, ok := l.userRatings[rf.IncludeUserRatingFor]
		if ok {
			result.UserRating = userRating
		}
	}

	return result
}
