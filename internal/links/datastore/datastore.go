package datastore

import (
	"context"
	"testing"
	"time"

	"github.com/dkinzler/linkboards/internal/links/domain"

	"github.com/dkinzler/kit/errors"
	ctime "github.com/dkinzler/kit/time"

	"github.com/stretchr/testify/assert"
)

type increasingTimer struct {
	i int
}

func (t *increasingTimer) curr() time.Time {
	t.i += 1
	return time.Date(2022, 0, 0, 0, 0, t.i, 0, time.UTC)
}

// Tests any implementation of links/domain/LinkDataStore
func DatastoreTest(ds domain.LinkDataStore, t *testing.T) {
	a := assert.New(t)

	timer := increasingTimer{}
	ctime.TimeFunc = timer.curr

	links := []domain.Link{
		{BoardId: "1", LinkId: "1", CreatedTime: ctime.CurrTimeUnixNano()},
		{BoardId: "1", LinkId: "2", CreatedTime: ctime.CurrTimeUnixNano()},
		{BoardId: "1", LinkId: "3", CreatedTime: ctime.CurrTimeUnixNano()},
		{BoardId: "1", LinkId: "4", CreatedTime: ctime.CurrTimeUnixNano()},
		{BoardId: "2", LinkId: "5", CreatedTime: ctime.CurrTimeUnixNano()},
	}

	userRatings := []struct {
		UserId  string
		BoardId string
		LinkId  string
		Rating  int
	}{
		{UserId: "1", BoardId: "1", LinkId: "4", Rating: 1},
		{UserId: "1", BoardId: "1", LinkId: "2", Rating: 1},
		{UserId: "2", BoardId: "1", LinkId: "2", Rating: 1},
		{UserId: "3", BoardId: "1", LinkId: "2", Rating: -1},
		{UserId: "4", BoardId: "1", LinkId: "2", Rating: 1},
		{UserId: "1", BoardId: "1", LinkId: "3", Rating: 1},
		{UserId: "2", BoardId: "1", LinkId: "3", Rating: -1},
		{UserId: "3", BoardId: "1", LinkId: "3", Rating: 1},
		{UserId: "4", BoardId: "1", LinkId: "3", Rating: -1},
		{UserId: "5", BoardId: "1", LinkId: "3", Rating: 1},
		{UserId: "6", BoardId: "1", LinkId: "3", Rating: 1},
		{UserId: "7", BoardId: "1", LinkId: "3", Rating: 1},
		{UserId: "1", BoardId: "2", LinkId: "5", Rating: 1},
	}

	ctx, cancel := getContext()
	defer cancel()

	for _, link := range links {
		err := ds.CreateLink(ctx, link.BoardId, link)
		a.Nil(err)
	}

	for _, userRating := range userRatings {
		err := ds.UpdateRating(ctx, userRating.BoardId, userRating.LinkId, domain.UserLinkRating{
			UserId: userRating.UserId,
			Rating: userRating.Rating,
		})
		a.Nil(err)
	}

	link, err := ds.Link(ctx, "1", "3", domain.LinkReturnFields{
		IncludeRating:        true,
		IncludeUserRatingFor: "4",
	})
	a.Nil(err)
	a.Equal(links[2], link.Link)
	a.Equal(domain.Rating{
		Score:     3,
		Upvotes:   5,
		Downvotes: -2,
	}, link.Rating)
	a.Equal(domain.UserLinkRating{
		UserId: "4",
		Rating: -1,
	}, link.UserRating)

	result, err := ds.Links(ctx, "1", domain.LinkReturnFields{
		IncludeRating:        true,
		IncludeUserRatingFor: "1",
	}, domain.NewLinkQueryParams().WithLimit(3).SortByNewest())
	a.Nil(err)
	a.Len(result, 3)
	// Links are sorted by newest
	a.Equal("4", result[0].Link.LinkId)
	a.Equal("3", result[1].Link.LinkId)
	a.Equal("2", result[2].Link.LinkId)

	result, err = ds.Links(ctx, "1", domain.LinkReturnFields{
		IncludeRating:        true,
		IncludeUserRatingFor: "1",
	}, domain.NewLinkQueryParams().WithLimit(5).SortByTop())
	a.Nil(err)
	a.Len(result, 4)
	// Links are sorted by score
	a.Equal("3", result[0].Link.LinkId)
	a.Equal("2", result[1].Link.LinkId)
	a.Equal("4", result[2].Link.LinkId)
	a.Equal("1", result[3].Link.LinkId)

	a.Equal(domain.LinkWithRating{
		Link: links[2],
		Rating: domain.Rating{
			Score:     3,
			Upvotes:   5,
			Downvotes: -2,
		},
		UserRating: domain.UserLinkRating{
			UserId: "1",
			Rating: 1,
		},
	}, result[0])

	err = ds.UpdateRating(ctx, "1", "3", domain.UserLinkRating{
		UserId: "10",
		Rating: -1,
	})
	a.Nil(err)
	err = ds.UpdateRating(ctx, "1", "3", domain.UserLinkRating{
		UserId: "11",
		Rating: -1,
	})
	a.Nil(err)
	err = ds.UpdateRating(ctx, "1", "4", domain.UserLinkRating{
		UserId: "12",
		Rating: 1,
	})
	a.Nil(err)
	err = ds.UpdateRating(ctx, "1", "2", domain.UserLinkRating{
		UserId: "13",
		Rating: 1,
	})
	a.Nil(err)

	// after all the updates the scores should be as follows
	// Link "3": 1
	// Link "4": 2
	// Link "2": 3
	result, err = ds.Links(ctx, "1", domain.LinkReturnFields{
		IncludeRating:        true,
		IncludeUserRatingFor: "1",
	}, domain.NewLinkQueryParams().WithLimit(5).SortByTop())
	a.Nil(err)
	a.Len(result, 4)
	// Links are sorted by score
	a.Equal("2", result[0].Link.LinkId)
	a.Equal("4", result[1].Link.LinkId)
	a.Equal("3", result[2].Link.LinkId)
	a.Equal("1", result[3].Link.LinkId)

	err = ds.DeleteLink(ctx, "1", "3")
	a.Nil(err)
	link, err = ds.Link(ctx, "1", "3", domain.LinkReturnFields{
		IncludeRating:        true,
		IncludeUserRatingFor: "1",
	})
	a.NotNil(err)
	a.True(errors.IsNotFoundError(err))
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}
