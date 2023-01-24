// generated code, do not modify
package transport

import (
	"context"
	"net/http"

	t "github.com/dkinzler/kit/transport/http"
	application "github.com/dkinzler/linkboards/internal/boards/application"
	kithttp "github.com/go-kit/kit/transport/http"
	mux "github.com/gorilla/mux"
)

func decodeHttpCreateBoardRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var nb application.NewBoard
	err := t.DecodeJSONBody(r, &nb)
	if err != nil {
		return nil, err
	}

	return CreateBoardRequest{Nb: nb}, nil
}

func decodeHttpDeleteBoardRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	return DeleteBoardRequest{BoardId: boardId}, nil
}

func decodeHttpEditBoardRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	var be application.BoardEdit
	err = t.DecodeJSONBody(r, &be)
	if err != nil {
		return nil, err
	}

	return EditBoardRequest{
		Be:      be,
		BoardId: boardId,
	}, nil
}

func decodeHttpBoardRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	return BoardRequest{BoardId: boardId}, nil
}

func decodeHttpBoardsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var qp application.QueryParams
	err := t.DecodeQueryParameters(r, &qp)
	if err != nil {
		return nil, err
	}

	return BoardsRequest{Qp: qp}, nil
}

func decodeHttpCreateInviteRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	var ni application.NewInvite
	err = t.DecodeJSONBody(r, &ni)
	if err != nil {
		return nil, err
	}

	return CreateInviteRequest{
		BoardId: boardId,
		Ni:      ni,
	}, nil
}

func decodeHttpRespondToInviteRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	inviteId, err := t.DecodeURLParameter(r, "inviteId")
	if err != nil {
		return nil, err
	}

	var ir application.InviteResponse
	err = t.DecodeJSONBody(r, &ir)
	if err != nil {
		return nil, err
	}

	return RespondToInviteRequest{
		BoardId:  boardId,
		InviteId: inviteId,
		Ir:       ir,
	}, nil
}

func decodeHttpDeleteInviteRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	inviteId, err := t.DecodeURLParameter(r, "inviteId")
	if err != nil {
		return nil, err
	}

	return DeleteInviteRequest{
		BoardId:  boardId,
		InviteId: inviteId,
	}, nil
}

func decodeHttpInvitesRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var qp application.QueryParams
	err := t.DecodeQueryParameters(r, &qp)
	if err != nil {
		return nil, err
	}

	return InvitesRequest{Qp: qp}, nil
}

func decodeHttpRemoveUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	userId, err := t.DecodeURLParameter(r, "userId")
	if err != nil {
		return nil, err
	}

	return RemoveUserRequest{
		BoardId: boardId,
		UserId:  userId,
	}, nil
}

func decodeHttpEditBoardUserRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	boardId, err := t.DecodeURLParameter(r, "boardId")
	if err != nil {
		return nil, err
	}

	userId, err := t.DecodeURLParameter(r, "userId")
	if err != nil {
		return nil, err
	}

	var bue application.BoardUserEdit
	err = t.DecodeJSONBody(r, &bue)
	if err != nil {
		return nil, err
	}

	return EditBoardUserRequest{
		BoardId: boardId,
		Bue:     bue,
		UserId:  userId,
	}, nil
}

func RegisterHttpHandlers(endpoints EndpointSet, router *mux.Router, opts []kithttp.ServerOption) {
	createBoardHandler := kithttp.NewServer(endpoints.CreateBoardEndpoint, decodeHttpCreateBoardRequest, t.MakeGenericJSONEncodeFunc(201), opts...)
	router.Handle("/boards", createBoardHandler).Methods("POST", "OPTIONS")

	boardsHandler := kithttp.NewServer(endpoints.BoardsEndpoint, decodeHttpBoardsRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards", boardsHandler).Methods("GET", "OPTIONS")

	deleteBoardHandler := kithttp.NewServer(endpoints.DeleteBoardEndpoint, decodeHttpDeleteBoardRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}", deleteBoardHandler).Methods("DELETE", "OPTIONS")

	editBoardHandler := kithttp.NewServer(endpoints.EditBoardEndpoint, decodeHttpEditBoardRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}", editBoardHandler).Methods("PATCH", "OPTIONS")

	boardHandler := kithttp.NewServer(endpoints.BoardEndpoint, decodeHttpBoardRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}", boardHandler).Methods("GET", "OPTIONS")

	createInviteHandler := kithttp.NewServer(endpoints.CreateInviteEndpoint, decodeHttpCreateInviteRequest, t.MakeGenericJSONEncodeFunc(201), opts...)
	router.Handle("/boards/{boardId}/invites", createInviteHandler).Methods("POST", "OPTIONS")

	respondToInviteHandler := kithttp.NewServer(endpoints.RespondToInviteEndpoint, decodeHttpRespondToInviteRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/invites/{inviteId}", respondToInviteHandler).Methods("POST", "OPTIONS")

	deleteInviteHandler := kithttp.NewServer(endpoints.DeleteInviteEndpoint, decodeHttpDeleteInviteRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/invites/{inviteId}", deleteInviteHandler).Methods("DELETE", "OPTIONS")

	removeUserHandler := kithttp.NewServer(endpoints.RemoveUserEndpoint, decodeHttpRemoveUserRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/users/{userId}", removeUserHandler).Methods("DELETE", "OPTIONS")

	editBoardUserHandler := kithttp.NewServer(endpoints.EditBoardUserEndpoint, decodeHttpEditBoardUserRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/boards/{boardId}/users/{userId}", editBoardUserHandler).Methods("PATCH", "OPTIONS")

	invitesHandler := kithttp.NewServer(endpoints.InvitesEndpoint, decodeHttpInvitesRequest, t.MakeGenericJSONEncodeFunc(200), opts...)
	router.Handle("/invites", invitesHandler).Methods("GET", "OPTIONS")
}
