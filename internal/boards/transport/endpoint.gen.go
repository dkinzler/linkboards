// generated code, do not modify
package transport

import (
	"context"

	e "github.com/d39b/kit/endpoint"
	application "github.com/d39b/linkboards/internal/boards/application"
	endpoint "github.com/go-kit/kit/endpoint"
)

type CreateBoardRequest struct {
	Nb application.NewBoard
}

func MakeCreateBoardEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateBoardRequest)
		r, err := svc.CreateBoard(ctx, req.Nb)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type DeleteBoardRequest struct {
	BoardId string
}

func MakeDeleteBoardEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteBoardRequest)
		err := svc.DeleteBoard(ctx, req.BoardId)
		return e.Response{
			Err: err,
			R:   nil,
		}, nil
	}
}

type EditBoardRequest struct {
	BoardId string
	Be      application.BoardEdit
}

func MakeEditBoardEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(EditBoardRequest)
		r, err := svc.EditBoard(ctx, req.BoardId, req.Be)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type BoardRequest struct {
	BoardId string
}

func MakeBoardEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(BoardRequest)
		r, err := svc.Board(ctx, req.BoardId)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type BoardsRequest struct {
	Qp application.QueryParams
}

func MakeBoardsEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(BoardsRequest)
		r, err := svc.Boards(ctx, req.Qp)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type CreateInviteRequest struct {
	BoardId string
	Ni      application.NewInvite
}

func MakeCreateInviteEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateInviteRequest)
		r, err := svc.CreateInvite(ctx, req.BoardId, req.Ni)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type RespondToInviteRequest struct {
	BoardId  string
	InviteId string
	Ir       application.InviteResponse
}

func MakeRespondToInviteEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RespondToInviteRequest)
		err := svc.RespondToInvite(ctx, req.BoardId, req.InviteId, req.Ir)
		return e.Response{
			Err: err,
			R:   nil,
		}, nil
	}
}

type DeleteInviteRequest struct {
	BoardId  string
	InviteId string
}

func MakeDeleteInviteEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DeleteInviteRequest)
		err := svc.DeleteInvite(ctx, req.BoardId, req.InviteId)
		return e.Response{
			Err: err,
			R:   nil,
		}, nil
	}
}

type InvitesRequest struct {
	Qp application.QueryParams
}

func MakeInvitesEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(InvitesRequest)
		r, err := svc.Invites(ctx, req.Qp)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type RemoveUserRequest struct {
	BoardId string
	UserId  string
}

func MakeRemoveUserEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RemoveUserRequest)
		err := svc.RemoveUser(ctx, req.BoardId, req.UserId)
		return e.Response{
			Err: err,
			R:   nil,
		}, nil
	}
}

type EditBoardUserRequest struct {
	BoardId string
	UserId  string
	Bue     application.BoardUserEdit
}

func MakeEditBoardUserEndpoint(svc application.BoardApplicationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(EditBoardUserRequest)
		r, err := svc.EditBoardUser(ctx, req.BoardId, req.UserId, req.Bue)
		return e.Response{
			Err: err,
			R:   r,
		}, nil
	}
}

type EndpointSet struct {
	CreateBoardEndpoint     endpoint.Endpoint
	DeleteBoardEndpoint     endpoint.Endpoint
	EditBoardEndpoint       endpoint.Endpoint
	BoardEndpoint           endpoint.Endpoint
	BoardsEndpoint          endpoint.Endpoint
	CreateInviteEndpoint    endpoint.Endpoint
	RespondToInviteEndpoint endpoint.Endpoint
	DeleteInviteEndpoint    endpoint.Endpoint
	InvitesEndpoint         endpoint.Endpoint
	RemoveUserEndpoint      endpoint.Endpoint
	EditBoardUserEndpoint   endpoint.Endpoint
}

type Middlewares struct {
	CreateBoardEndpoint     []endpoint.Middleware
	DeleteBoardEndpoint     []endpoint.Middleware
	EditBoardEndpoint       []endpoint.Middleware
	BoardEndpoint           []endpoint.Middleware
	BoardsEndpoint          []endpoint.Middleware
	CreateInviteEndpoint    []endpoint.Middleware
	RespondToInviteEndpoint []endpoint.Middleware
	DeleteInviteEndpoint    []endpoint.Middleware
	InvitesEndpoint         []endpoint.Middleware
	RemoveUserEndpoint      []endpoint.Middleware
	EditBoardUserEndpoint   []endpoint.Middleware
}

func NewEndpoints(svc application.BoardApplicationService, mws Middlewares) EndpointSet {
	var createBoardEndpoint endpoint.Endpoint
	{
		createBoardEndpoint = MakeCreateBoardEndpoint(svc)
		createBoardEndpoint = e.ApplyMiddlewares(createBoardEndpoint, mws.CreateBoardEndpoint...)
	}

	var deleteBoardEndpoint endpoint.Endpoint
	{
		deleteBoardEndpoint = MakeDeleteBoardEndpoint(svc)
		deleteBoardEndpoint = e.ApplyMiddlewares(deleteBoardEndpoint, mws.DeleteBoardEndpoint...)
	}

	var editBoardEndpoint endpoint.Endpoint
	{
		editBoardEndpoint = MakeEditBoardEndpoint(svc)
		editBoardEndpoint = e.ApplyMiddlewares(editBoardEndpoint, mws.EditBoardEndpoint...)
	}

	var boardEndpoint endpoint.Endpoint
	{
		boardEndpoint = MakeBoardEndpoint(svc)
		boardEndpoint = e.ApplyMiddlewares(boardEndpoint, mws.BoardEndpoint...)
	}

	var boardsEndpoint endpoint.Endpoint
	{
		boardsEndpoint = MakeBoardsEndpoint(svc)
		boardsEndpoint = e.ApplyMiddlewares(boardsEndpoint, mws.BoardsEndpoint...)
	}

	var createInviteEndpoint endpoint.Endpoint
	{
		createInviteEndpoint = MakeCreateInviteEndpoint(svc)
		createInviteEndpoint = e.ApplyMiddlewares(createInviteEndpoint, mws.CreateInviteEndpoint...)
	}

	var respondToInviteEndpoint endpoint.Endpoint
	{
		respondToInviteEndpoint = MakeRespondToInviteEndpoint(svc)
		respondToInviteEndpoint = e.ApplyMiddlewares(respondToInviteEndpoint, mws.RespondToInviteEndpoint...)
	}

	var deleteInviteEndpoint endpoint.Endpoint
	{
		deleteInviteEndpoint = MakeDeleteInviteEndpoint(svc)
		deleteInviteEndpoint = e.ApplyMiddlewares(deleteInviteEndpoint, mws.DeleteInviteEndpoint...)
	}

	var invitesEndpoint endpoint.Endpoint
	{
		invitesEndpoint = MakeInvitesEndpoint(svc)
		invitesEndpoint = e.ApplyMiddlewares(invitesEndpoint, mws.InvitesEndpoint...)
	}

	var removeUserEndpoint endpoint.Endpoint
	{
		removeUserEndpoint = MakeRemoveUserEndpoint(svc)
		removeUserEndpoint = e.ApplyMiddlewares(removeUserEndpoint, mws.RemoveUserEndpoint...)
	}

	var editBoardUserEndpoint endpoint.Endpoint
	{
		editBoardUserEndpoint = MakeEditBoardUserEndpoint(svc)
		editBoardUserEndpoint = e.ApplyMiddlewares(editBoardUserEndpoint, mws.EditBoardUserEndpoint...)
	}

	return EndpointSet{
		BoardEndpoint:           boardEndpoint,
		BoardsEndpoint:          boardsEndpoint,
		CreateBoardEndpoint:     createBoardEndpoint,
		CreateInviteEndpoint:    createInviteEndpoint,
		DeleteBoardEndpoint:     deleteBoardEndpoint,
		DeleteInviteEndpoint:    deleteInviteEndpoint,
		EditBoardEndpoint:       editBoardEndpoint,
		EditBoardUserEndpoint:   editBoardUserEndpoint,
		InvitesEndpoint:         invitesEndpoint,
		RemoveUserEndpoint:      removeUserEndpoint,
		RespondToInviteEndpoint: respondToInviteEndpoint,
	}
}
