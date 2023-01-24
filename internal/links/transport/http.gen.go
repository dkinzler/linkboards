// generated code, do not modify
package transport

import (
	"context"
	"net/http"

	application "github.com/dkinzler/linkboards/internal/links/application"

	t "github.com/dkinzler/kit/transport/http"
	kithttp "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
)

func decodeHttpCreateLinkRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	var nl application.NewLink
	err = t.DecodeJSONBody(r, &nl)
	if err != nil {
		return nil, err
	}

	return CreateLinkRequest{
		BoardId: boardId,
		Nl:      nl,
	}, nil
}

func decodeHttpDeleteLinkRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	linkId, err := t.DecodeURLParameter(r, "linkId")
	if err != nil {
		return nil, err
	}

	return DeleteLinkRequest{
		BoardId: boardId,
		LinkId:  linkId,
	}, nil
}

func decodeHttpRateLinkRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	linkId, err := t.DecodeURLParameter(r, "linkId")
	if err != nil {
		return nil, err
	}

	var lr application.LinkRating
	err = t.DecodeJSONBody(r, &lr)
	if err != nil {
		return nil, err
	}

	return RateLinkRequest{
		BoardId: boardId,
		LinkId:  linkId,
		Lr:      lr,
	}, nil
}

func decodeHttpLinkRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	linkId, err := t.DecodeURLParameter(r, "linkId")
	if err != nil {
		return nil, err
	}

	return LinkRequest{
		BoardId: boardId,
		LinkId:  linkId,
	}, nil
}

func decodeHttpLinksRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	var qp application.LinkQueryParams
	err = t.DecodeQueryParameters(r, &qp)
	if err != nil {
		return nil, err
	}

	return LinksRequest{
		BoardId: boardId,
		Qp:      qp,
	}, nil
}

func RegisterHttpHandlers(endpoints EndpointSet, router *mux.Router, opts []kithttp.ServerOption) {
	createLinkHandler := kithttp.NewServer(endpoints.CreateLinkEndpoint, decodeHttpCreateLinkRequest, t.MakeGenericJSONEncodeFunc(201), opts...)
	router.Handle("/boards/{boardId}/links", createLinkHandler).Methods("POST", "OPTIONS")

	linksHandler := kithttp.NewServer(endpoints.LinksEndpoint, decodeHttpLinksRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/links", linksHandler).Methods("GET", "OPTIONS")

	deleteLinkHandler := kithttp.NewServer(endpoints.DeleteLinkEndpoint, decodeHttpDeleteLinkRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/links/{linkId}", deleteLinkHandler).Methods("DELETE", "OPTIONS")

	linkHandler := kithttp.NewServer(endpoints.LinkEndpoint, decodeHttpLinkRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/links/{linkId}", linkHandler).Methods("GET", "OPTIONS")

	rateLinkHandler := kithttp.NewServer(endpoints.RateLinkEndpoint, decodeHttpRateLinkRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/links/{linkId}/ratings", rateLinkHandler).Methods("POST", "OPTIONS")
}
