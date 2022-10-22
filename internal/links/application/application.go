package application

import (
	"context"
	"go-sample/internal/auth"
	"go-sample/internal/links/domain"

	"github.com/d39b/kit/errors"
)

// @Kit{"endpointPackage":"internal/links/transport", "httpPackage":"internal/links/transport"}
type LinkApplicationService interface {
	// @Kit{
	//	"httpParams": ["url", "json"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/links", "method":"POST", "successCode": 201}}]
	// }
	CreateLink(ctx context.Context, boardId string, nl NewLink) (Link, error)
	// @Kit{
	//	"httpParams": ["url", "url"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/links/{linkId}", "method":"DELETE"}}]
	// }
	DeleteLink(ctx context.Context, boardId string, linkId string) error
	// @Kit{
	//	"httpParams": ["url", "url", "json"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/links/{linkId}/ratings", "method":"POST"}}]
	// }
	RateLink(ctx context.Context, boardId string, linkId string, lr LinkRating) error
	// @Kit{
	//	"httpParams": ["url", "url"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/links/{linkId}", "method":"GET"}}]
	// }
	Link(ctx context.Context, boardId string, linkId string) (Link, error)
	// @Kit{
	//	"httpParams": ["url", "query"],
	//	"endpoints": [{"http": {"path":"/boards/{boardId}/links", "method":"GET"}}]
	// }
	Links(ctx context.Context, boardId string, qp LinkQueryParams) ([]Link, error)
}

type NewLink struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type Link struct {
	BoardId string `json:"boardId"`
	LinkId  string `json:"linkId"`

	Title string `json:"title"`
	Url   string `json:"url"`

	CreatedTime int64       `json:"createdTime"`
	CreatedBy   domain.User `json:"createdBy"`

	Score     int `json:"score"`
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`

	// Rating of the user making a request
	UserRating int `json:"userRating"`
}

func linkFromDomainLink(l domain.LinkWithRating) Link {
	return Link{
		BoardId:     l.Link.BoardId,
		LinkId:      l.Link.LinkId,
		Title:       l.Link.Title,
		Url:         l.Link.Url,
		CreatedTime: l.Link.CreatedTime,
		CreatedBy:   l.Link.CreatedBy,
		Score:       l.Rating.Score,
		Upvotes:     l.Rating.Upvotes,
		Downvotes:   l.Rating.Downvotes,
		UserRating:  l.UserRating.Rating,
	}
}

type LinkRating struct {
	// Valid values are +1 for an upvote and -1 for a downvote
	Rating int `json:"rating"`
}

type LinkQueryParams struct {
	// Number of results to return.
	// Must be between 10 and 100, defaults to 20.
	Limit int
	// Valid values are "newest" and "top", defaults to "newest".
	Sort              string
	CursorScore       *int
	CursorCreatedTime *int64
}

func (qp LinkQueryParams) toDatastoreQueryParams() domain.LinkQueryParams {
	// has defaults already set, limit=20, sort=newest
	q := domain.NewLinkQueryParams()
	if qp.Limit >= 10 && qp.Limit <= 100 {
		q = q.WithLimit(qp.Limit)
	} else {
		q = q.WithLimit(20)
	}

	if qp.Sort == "top" {
		q = q.SortByTop()
	}

	if qp.CursorScore != nil {
		q = q.WithScoreCursor(*qp.CursorScore)
	}

	if qp.CursorCreatedTime != nil {
		q = q.WithCreatedTimeCursor(*qp.CursorCreatedTime)
	}

	return q
}

type LinkQueryResult struct {
	Result            []Link `json:"result"`
	CursorScore       *int   `json:"cursorScore"`
	CursorCreatedTime *int64 `json:"cursorCreatedTime"`
}

type linkApplicationService struct {
	linkService   *domain.LinkService
	linkDataStore domain.LinkDataStore
	authChecker   *auth.BoardAuthorizationChecker
}

func NewLinkApplicationService(linkDataStore domain.LinkDataStore, authorizationStore auth.AuthorizationStore) LinkApplicationService {
	return &linkApplicationService{
		linkService:   domain.NewLinkService(linkDataStore),
		linkDataStore: linkDataStore,
		authChecker:   NewAuthorizationChecker(authorizationStore),
	}
}

func newServiceError(inner error, code errors.ErrorCode) errors.Error {
	return errors.New(inner, "LinkApplicationService", code)
}

func newUnauthenticatedError() errors.Error {
	return newServiceError(nil, errors.Unauthenticated)
}

func newPermissionDeniedError() errors.Error {
	return newServiceError(nil, errors.PermissionDenied)
}

func (svc *linkApplicationService) CreateLink(ctx context.Context, boardId string, nl NewLink) (Link, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return Link{}, newUnauthenticatedError()
	}

	az, err := svc.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return Link{}, err
	}

	if !az.HasScope(createLinkScope) {
		return Link{}, newPermissionDeniedError()
	}

	link, err := svc.linkService.CreateLink(ctx, boardId, nl.Title, nl.Url, toDomainUser(user))
	return Link{
		BoardId:     link.BoardId,
		LinkId:      link.LinkId,
		Title:       link.Title,
		Url:         link.Url,
		CreatedTime: link.CreatedTime,
		CreatedBy:   link.CreatedBy,
		Score:       0,
		Downvotes:   0,
		Upvotes:     0,
		UserRating:  0,
	}, err
}

func (svc *linkApplicationService) DeleteLink(ctx context.Context, boardId string, linkId string) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return newUnauthenticatedError()
	}

	// user that created the link can also delete it
	link, err := svc.linkDataStore.Link(ctx, boardId, linkId, domain.LinkReturnFields{})
	if err != nil {
		if errors.IsNotFoundError(err) {
			return newServiceError(err, errors.NotFound)
		}
		return newServiceError(err, errors.Internal)
	}

	// TODO It could happen that the user is no longer a member of the board the link is in.
	// Should they then be able to delte the link or not?

	// if the user that made the request is not the user that created the link
	// they need the delete link scope
	if link.Link.CreatedBy.UserId != user.UserId {
		az, err := svc.authChecker.GetAuthorization(ctx, boardId, user.UserId)
		if err != nil {
			return err
		}

		if !az.HasScope(deleteLinkScope) {
			return newPermissionDeniedError()
		}
	}

	err = svc.linkDataStore.DeleteLink(ctx, boardId, linkId)
	if err != nil {
		return newServiceError(nil, errors.Internal).WithInternalMessage("could not delete link")
	}

	return nil
}

func (svc *linkApplicationService) RateLink(ctx context.Context, boardId string, linkId string, lr LinkRating) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return newUnauthenticatedError()
	}

	az, err := svc.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return err
	}

	if !az.HasScope(rateLinkScope) {
		return newPermissionDeniedError()
	}

	_, err = svc.linkService.UpdateUserRating(ctx, boardId, linkId, lr.Rating, toDomainUser(user))
	return err
}

func (svc *linkApplicationService) Link(ctx context.Context, boardId string, linkId string) (Link, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return Link{}, newUnauthenticatedError()
	}

	az, err := svc.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return Link{}, err
	}

	if !az.HasScope(queryLinksScope) {
		return Link{}, newPermissionDeniedError()
	}

	link, err := svc.linkDataStore.Link(ctx, boardId, linkId, domain.LinkReturnFields{
		IncludeRating:        true,
		IncludeUserRatingFor: user.UserId,
	})
	if err != nil {
		if errors.IsNotFoundError(err) {
			return Link{}, newServiceError(err, errors.NotFound)
		}
		return Link{}, newServiceError(err, errors.Internal)
	}

	return linkFromDomainLink(link), nil
}

func (svc *linkApplicationService) Links(ctx context.Context, boardId string, qp LinkQueryParams) ([]Link, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, newUnauthenticatedError()
	}

	az, err := svc.authChecker.GetAuthorization(ctx, boardId, user.UserId)
	if err != nil {
		return nil, err
	}

	if !az.HasScope(queryLinksScope) {
		return nil, newPermissionDeniedError()
	}

	links, err := svc.linkDataStore.Links(ctx, boardId, domain.LinkReturnFields{
		IncludeRating:        true,
		IncludeUserRatingFor: user.UserId,
	}, qp.toDatastoreQueryParams())
	if err != nil {
		return nil, newServiceError(err, errors.Internal)
	}

	result := make([]Link, len(links))
	for i, link := range links {
		result[i] = linkFromDomainLink(link)
	}

	return result, nil
}
