// generated code, do not modify
package transport

import (
	"context"
	application "linkboards/internal/links/application"

	e "github.com/d39b/kit/endpoint"
	endpoint "github.com/go-kit/kit/endpoint"
)

type CreateLinkRequest struct {
	BoardId string
	Nl      application.NewLink
}

func MakeCreateLinkEndpoint(svc application.LinkApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateLinkRequest)
		r, err := svc.CreateLink(ctx, req.BoardId, req.Nl)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type DeleteLinkRequest struct {
	BoardId string
	LinkId  string
}

func MakeDeleteLinkEndpoint(svc application.LinkApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteLinkRequest)
		err := svc.DeleteLink(ctx, req.BoardId, req.LinkId)
		return e.Response{
			Err: err,
			R:   nil,
		}, nil
	}
}

type RateLinkRequest struct {
	BoardId string
	LinkId  string
	Lr      application.LinkRating
}

func MakeRateLinkEndpoint(svc application.LinkApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RateLinkRequest)
		err := svc.RateLink(ctx, req.BoardId, req.LinkId, req.Lr)
		return e.Response{
			Err: err,
			R:   nil,
		}, nil
	}
}

type LinkRequest struct {
	BoardId string
	LinkId  string
}

func MakeLinkEndpoint(svc application.LinkApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(LinkRequest)
		r, err := svc.Link(ctx, req.BoardId, req.LinkId)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type LinksRequest struct {
	BoardId string
	Qp      application.LinkQueryParams
}

func MakeLinksEndpoint(svc application.LinkApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(LinksRequest)
		r, err := svc.Links(ctx, req.BoardId, req.Qp)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type EndpointSet struct {
	CreateLinkEndpoint endpoint.Endpoint
	DeleteLinkEndpoint endpoint.Endpoint
	RateLinkEndpoint   endpoint.Endpoint
	LinkEndpoint       endpoint.Endpoint
	LinksEndpoint      endpoint.Endpoint
}

type Middlewares struct {
	CreateLinkEndpoint []endpoint.Middleware
	DeleteLinkEndpoint []endpoint.Middleware
	RateLinkEndpoint   []endpoint.Middleware
	LinkEndpoint       []endpoint.Middleware
	LinksEndpoint      []endpoint.Middleware
}

func NewEndpoints(svc application.LinkApplicationService, mws Middlewares) EndpointSet {
	var createLinkEndpoint endpoint.Endpoint
	{
		createLinkEndpoint = MakeCreateLinkEndpoint(svc)
		createLinkEndpoint = e.ApplyMiddlewares(createLinkEndpoint, mws.CreateLinkEndpoint...)
	}

	var deleteLinkEndpoint endpoint.Endpoint
	{
		deleteLinkEndpoint = MakeDeleteLinkEndpoint(svc)
		deleteLinkEndpoint = e.ApplyMiddlewares(deleteLinkEndpoint, mws.DeleteLinkEndpoint...)
	}

	var rateLinkEndpoint endpoint.Endpoint
	{
		rateLinkEndpoint = MakeRateLinkEndpoint(svc)
		rateLinkEndpoint = e.ApplyMiddlewares(rateLinkEndpoint, mws.RateLinkEndpoint...)
	}

	var linkEndpoint endpoint.Endpoint
	{
		linkEndpoint = MakeLinkEndpoint(svc)
		linkEndpoint = e.ApplyMiddlewares(linkEndpoint, mws.LinkEndpoint...)
	}

	var linksEndpoint endpoint.Endpoint
	{
		linksEndpoint = MakeLinksEndpoint(svc)
		linksEndpoint = e.ApplyMiddlewares(linksEndpoint, mws.LinksEndpoint...)
	}

	return EndpointSet{
		CreateLinkEndpoint: createLinkEndpoint,
		DeleteLinkEndpoint: deleteLinkEndpoint,
		LinkEndpoint:       linkEndpoint,
		LinksEndpoint:      linksEndpoint,
		RateLinkEndpoint:   rateLinkEndpoint,
	}
}
