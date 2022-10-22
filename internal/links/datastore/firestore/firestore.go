package firestore

import (
	"context"
	"linkboards/internal/links/domain"

	"github.com/d39b/kit/errors"
	fs "github.com/d39b/kit/firebase/firestore"

	"cloud.google.com/go/firestore"
)

const linksCollectionName = "links"

type FirestoreLinkDataStore struct {
	client          *firestore.Client
	linksCollection *firestore.CollectionRef
}

func NewFirestoreLinkDataStore(client *firestore.Client) *FirestoreLinkDataStore {
	return &FirestoreLinkDataStore{
		client:          client,
		linksCollection: client.Collection(linksCollectionName),
	}
}

// Since the number of users of a board is limited, so is the number of ratings for a link.
// Therefore we can store a link together with its rating aggregate and all the user ratings in a single document.
// Also we don't have to worry about contention, it is very unlikely that there are multiple votes happening for the same link
// within the same second (firestore doesn't support write rates above 1 write/s for a single document for longer durations).
//
// We could also create custom firestore types for domain.Link, domain.Rating etc.
// This would clarify and make explicit the names of some fields.
// E.g. when we query the top links, we will order by the field "rating.Score", because
// the domain.Rating struct doesn't have a firestore tag, but we know that the firestore package
// will just use the struct field name (i.e. Score).
// However defining custom types introduces alot of boilerplate code.
// Also defining firestore tags on domain types like domain.Rating would leak implementation details
// from the firestore package into the domain package.
// All in all for a simple project like this it is probably fine to just use the domain types directly.
// In a bigger and more complex project one could use custom types and maybe use code generation to create them.
type fsLink struct {
	BoardId     string                           `firestore:"boardId"`
	LinkId      string                           `firestore:"linkId"`
	Link        domain.Link                      `firestore:"link"`
	Rating      domain.Rating                    `firestore:"rating"`
	UserRatings map[string]domain.UserLinkRating `firestore:"userRatings"`
}

func domainLinkFromFirestoreLink(l fsLink, rf domain.LinkReturnFields) domain.LinkWithRating {
	result := domain.LinkWithRating{
		Link: l.Link,
	}

	if rf.IncludeRating {
		result.Rating = l.Rating
	}

	if rf.IncludeUserRatingFor != "" {
		userRating, ok := l.UserRatings[rf.IncludeUserRatingFor]
		if ok {
			result.UserRating = userRating
		}
	}

	return result
}

func (ds *FirestoreLinkDataStore) CreateLink(ctx context.Context, boardId string, link domain.Link) error {
	fsLink := fsLink{
		BoardId: boardId,
		LinkId:  link.LinkId,
		Link:    link,
		Rating: domain.Rating{
			Score:     0,
			Upvotes:   0,
			Downvotes: 0,
		},
		UserRatings: map[string]domain.UserLinkRating{},
	}

	err := fs.CreateDocument(ctx, ds.linksCollection, link.LinkId, fsLink)
	if err != nil {
		return err
	}
	return nil
}

func (ds *FirestoreLinkDataStore) DeleteLink(ctx context.Context, boardId string, linkId string) error {
	err := fs.DeleteDocument(ctx, ds.linksCollection, linkId)
	return err
}

func (ds *FirestoreLinkDataStore) UpdateRating(ctx context.Context, boardId string, linkId string, userRating domain.UserLinkRating) error {
	err := ds.client.RunTransaction(ctx, func(c context.Context, t *firestore.Transaction) error {
		snap, err := t.Get(ds.linksCollection.Doc(linkId))
		if err != nil {
			return fs.ParseFirestoreError(err)
		}
		var link fsLink
		err = fs.UnmarshalDocSnapshot(snap, &link)
		if err != nil {
			return err
		}

		newRating := link.Rating
		if oldUserRating, ok := link.UserRatings[userRating.UserId]; ok {
			if oldUserRating.Rating > 0 {
				newRating.Upvotes -= oldUserRating.Rating
			} else {
				newRating.Downvotes -= oldUserRating.Rating
			}
		}
		if userRating.Rating > 0 {
			newRating.Upvotes += userRating.Rating
		} else {
			newRating.Downvotes += userRating.Rating
		}
		newRating.Score = newRating.Upvotes + newRating.Downvotes

		updates := []firestore.Update{
			{Path: "userRatings." + userRating.UserId, Value: userRating},
			{Path: "rating", Value: newRating},
		}
		err = t.Update(ds.linksCollection.Doc(linkId), updates)
		if err != nil {
			return fs.NewFirestoreError(err, errors.Internal)
		}

		return nil
	}, firestore.MaxAttempts(1))
	return err
}

func (ds *FirestoreLinkDataStore) Link(ctx context.Context, boardId string, linkId string, rf domain.LinkReturnFields) (domain.LinkWithRating, error) {
	var link fsLink
	err := fs.GetDocumentById(ctx, ds.linksCollection, linkId, &link)
	if err != nil {
		return domain.LinkWithRating{}, err
	}

	return domainLinkFromFirestoreLink(link, rf), nil
}

func (ds *FirestoreLinkDataStore) Links(ctx context.Context, boardId string, rf domain.LinkReturnFields, qp domain.LinkQueryParams) ([]domain.LinkWithRating, error) {
	query := ds.linksCollection.Where("boardId", "==", boardId)

	pathsToSelect := []string{"boardId", "linkId", "link"}
	if rf.IncludeRating {
		pathsToSelect = append(pathsToSelect, "rating")
	}
	if rf.IncludeUserRatingFor != "" {
		pathsToSelect = append(pathsToSelect, "userRatings."+rf.IncludeUserRatingFor)
	}
	query = query.Select(pathsToSelect...)

	if qp.SortOrder == domain.SortOrderNewest {
		query = query.OrderBy("link.CreatedTime", firestore.Desc)
		if qp.CursorCreatedTime != 0 {
			query = query.StartAt(qp.CursorCreatedTime)
		}
	} else if qp.SortOrder == domain.SortOrderTop {
		query = query.OrderBy("rating.Score", firestore.Desc).OrderBy("link.CreatedTime", firestore.Desc)
		if qp.CursorScore != nil {
			query = query.StartAt(*qp.CursorScore, qp.CursorCreatedTime)
		}
	}

	if qp.Limit > 0 {
		limit := qp.Limit
		if limit > 100 {
			// don't let limit be too large
			limit = 100
		}
		query = query.Limit(limit)
	} else {
		query = query.Limit(20)
	}

	snaps, err := fs.GetDocumentsForQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]domain.LinkWithRating, len(snaps))
	for i, snap := range snaps {
		var link fsLink
		err = fs.UnmarshalDocSnapshot(snap, &link)
		if err != nil {
			return nil, err
		}
		result[i] = domainLinkFromFirestoreLink(link, rf)
	}

	return result, nil
}
