package client

import (
	"fmt"
	"net/url"
)

type ApiClient struct {
	UserId string
	Token  string

	Address string
}

func NewApiClient(userId, token, address string) *ApiClient {
	return &ApiClient{
		UserId:  userId,
		Token:   token,
		Address: address,
	}
}

func (a ApiClient) baseRequest() request {
	return request{
		BaseUrl: url.URL{
			Scheme: "http",
			Host:   a.Address,
		},
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %v", a.Token),
		},
	}
}

func (a ApiClient) CreateBoard(newBoard NewBoard) (Board, response, error) {
	r := a.baseRequest()
	r.Method = "POST"
	r.Path = "/boards"
	r.Body = newBoard

	var board Board
	resp, err := doRequest(r, &board)
	if err != nil {
		return board, resp, err
	}

	return board, resp, nil
}

func (a ApiClient) DeleteBoard(boardId string) (response, error) {
	r := a.baseRequest()
	r.Method = "DELETE"
	r.Path = fmt.Sprintf("/boards/%v", boardId)

	resp, err := doRequest(r, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (a ApiClient) EditBoard(boardId string, be BoardEdit) (Board, response, error) {
	r := a.baseRequest()
	r.Method = "PATCH"
	r.Path = fmt.Sprintf("/boards/%v", boardId)
	r.Body = be

	var board Board
	resp, err := doRequest(r, &board)
	if err != nil {
		return board, resp, err
	}

	return board, resp, nil
}

func (a ApiClient) GetBoard(boardId string) (Board, response, error) {
	r := a.baseRequest()
	r.Method = "GET"
	r.Path = fmt.Sprintf("/boards/%v", boardId)

	var board Board
	resp, err := doRequest(r, &board)
	if err != nil {
		return board, resp, err
	}

	return board, resp, nil
}

func (a ApiClient) GetBoards(qp QueryParams) ([]Board, response, error) {
	r := a.baseRequest()
	r.Method = "GET"
	r.Path = "/boards"
	uqp := map[string]string{}
	if qp.Cursor != 0 {
		uqp["cursor"] = fmt.Sprint(qp.Cursor)
	}
	if qp.Limit != 0 {
		uqp["limit"] = fmt.Sprint(qp.Limit)
	}
	r.QueryParams = uqp

	var boards []Board
	resp, err := doRequest(r, &boards)
	if err != nil {
		return boards, resp, err
	}

	return boards, resp, nil
}

func (a ApiClient) CreateInvite(boardId string, ni NewInvite) (BoardInvite, response, error) {
	r := a.baseRequest()
	r.Method = "POST"
	r.Path = fmt.Sprintf("/boards/%v/invites", boardId)
	r.Body = ni

	var invite BoardInvite
	resp, err := doRequest(r, &invite)
	if err != nil {
		return invite, resp, err
	}

	return invite, resp, nil
}

func (a ApiClient) RespondToInvite(boardId string, inviteId string, ir InviteResponse) (response, error) {
	r := a.baseRequest()
	r.Method = "POST"
	r.Path = fmt.Sprintf("/boards/%v/invites/%v", boardId, inviteId)
	r.Body = ir

	resp, err := doRequest(r, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (a ApiClient) DeleteInvite(boardId string, inviteId string) (response, error) {
	r := a.baseRequest()
	r.Method = "DELETE"
	r.Path = fmt.Sprintf("/boards/%v/invites/%v", boardId, inviteId)

	resp, err := doRequest(r, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (a ApiClient) GetInvites(qp QueryParams) ([]BoardInvite, response, error) {
	r := a.baseRequest()
	r.Method = "GET"
	r.Path = "/invites"
	uqp := map[string]string{}
	if qp.Cursor != 0 {
		uqp["cursor"] = fmt.Sprint(qp.Cursor)
	}
	if qp.Limit != 0 {
		uqp["limit"] = fmt.Sprint(qp.Limit)
	}
	r.QueryParams = uqp

	var invites []BoardInvite
	resp, err := doRequest(r, &invites)
	if err != nil {
		return invites, resp, err
	}

	return invites, resp, nil
}

func (a ApiClient) RemoveUserFromBoard(boardId string, userId string) (response, error) {
	r := a.baseRequest()
	r.Method = "DELETE"
	r.Path = fmt.Sprintf("/boards/%v/users/%v", boardId, userId)

	resp, err := doRequest(r, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (a ApiClient) EditBoardUser(boardId string, userId string, bue BoardUserEdit) (BoardUser, response, error) {
	r := a.baseRequest()
	r.Method = "PATCH"
	r.Path = fmt.Sprintf("/boards/%v/users/%v", boardId, userId)
	r.Body = bue

	var boardUser BoardUser
	resp, err := doRequest(r, &boardUser)
	if err != nil {
		return boardUser, resp, err
	}

	return boardUser, resp, nil
}

func (a ApiClient) CreateLink(boardId string, nl NewLink) (Link, response, error) {
	r := a.baseRequest()
	r.Method = "POST"
	r.Path = fmt.Sprintf("/boards/%v/links", boardId)
	r.Body = nl

	var link Link
	resp, err := doRequest(r, &link)
	if err != nil {
		return link, resp, err
	}

	return link, resp, nil
}

func (a ApiClient) DeleteLink(boardId string, linkId string) (response, error) {
	r := a.baseRequest()
	r.Method = "DELETE"
	r.Path = fmt.Sprintf("/boards/%v/links/%v", boardId, linkId)

	resp, err := doRequest(r, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (a ApiClient) RateLink(boardId string, linkId string, lr LinkRating) (response, error) {
	r := a.baseRequest()
	r.Method = "POST"
	r.Path = fmt.Sprintf("/boards/%v/links/%v/ratings", boardId, linkId)
	r.Body = lr

	resp, err := doRequest(r, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (a ApiClient) GetLink(boardId string, linkId string) (Link, response, error) {
	r := a.baseRequest()
	r.Method = "GET"
	r.Path = fmt.Sprintf("/boards/%v/links/%v", boardId, linkId)

	var link Link
	resp, err := doRequest(r, &link)
	if err != nil {
		return link, resp, err
	}

	return link, resp, nil
}

func (a ApiClient) GetLinks(boardId string, qp LinkQueryParams) ([]Link, response, error) {
	r := a.baseRequest()
	r.Method = "GET"
	r.Path = fmt.Sprintf("/boards/%v/links", boardId)
	uqp := map[string]string{}
	if qp.Limit != 0 {
		uqp["limit"] = fmt.Sprint(qp.Limit)
	}
	if qp.Sort != "" {
		uqp["sort"] = qp.Sort
	}
	if qp.CursorCreatedTime != nil {
		uqp["cursorCreatedTime"] = fmt.Sprint(*qp.CursorCreatedTime)
	}
	if qp.CursorScore != nil {
		uqp["cursorScore"] = fmt.Sprint(*qp.CursorScore)
	}
	r.QueryParams = uqp

	var links []Link
	resp, err := doRequest(r, &links)
	if err != nil {
		return links, resp, err
	}

	return links, resp, nil
}
