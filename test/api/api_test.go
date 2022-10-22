package api

import (
	"errors"
	"fmt"
	"go-sample/tools/client"
	"log"
	"math/rand"
	"os"
	"sort"
	"testing"

	"github.com/d39b/kit/firebase/emulator"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Skip these tests if the environment variable is not set.
	if pid := os.Getenv("FIREBASE_PROJECT_ID"); pid == "" {
		log.Println("set FIREBASE_PROJECT_ID to run these tests")
		os.Exit(0)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

func newClientBuilder() (*clientBuilder, error) {
	authClient, err := emulator.NewAuthEmulatorClient()
	if err != nil {
		return nil, err
	}

	apiAddress := "localhost:9001"
	if aa := os.Getenv("API_ADDRESS"); aa != "" {
		apiAddress = aa
	}

	ub := &clientBuilder{
		authClient: authClient,
		address:    apiAddress,
	}

	return ub, nil
}

func TestEditBoard(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)

	user1, client1, err := cb.createUser()
	a.Nil(err)
	user2, client2, err := cb.createUser()
	a.Nil(err)

	board, resp, err := client1.CreateBoard(client.NewBoard{
		Name:        "test board 123",
		Description: "abc",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	a.Equal("test board 123", board.Name)
	a.Equal("abc", board.Description)

	boardId := board.BoardId

	// edit board
	newBoard, resp, err := client1.EditBoard(boardId, client.BoardEdit{}.WithName("new board name").WithDescription("new board description"))
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(boardId, newBoard.BoardId)
	a.Equal("new board name", newBoard.Name)
	a.Equal("new board description", newBoard.Description)
	a.Equal(user1.Uid, newBoard.ModifiedBy.UserId)

	newBoard, resp, err = client1.GetBoard(boardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(boardId, newBoard.BoardId)
	a.Equal("new board name", newBoard.Name)
	a.Equal("new board description", newBoard.Description)
	a.Equal(user1.Uid, newBoard.ModifiedBy.UserId)

	// user2 is not part of the board, should not be able to edit board
	newBoard, resp, err = client2.EditBoard(boardId, client.BoardEdit{}.WithName("another board name").WithDescription("xyz"))
	a.Nil(err)
	a.Equal(403, resp.HttpCode)
	a.Empty(newBoard)

	// add the user to the board with editor role, they should then be able to edit the board
	invite, resp, err := client1.CreateInvite(boardId, client.NewInvite{
		User: client.User{UserId: user2.Uid},
		Role: client.RoleEditor,
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	// accept the invite
	resp, err = client2.RespondToInvite(boardId, invite.InviteId, client.InviteResponse{Response: client.InviteResponseAccept})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	// user2 is not part of the board, should not be able to edit board
	newBoard, resp, err = client2.EditBoard(boardId, client.BoardEdit{}.WithName("another board name").WithDescription("xyz"))
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(boardId, newBoard.BoardId)
	a.Equal("another board name", newBoard.Name)
	a.Equal("xyz", newBoard.Description)
}

func TestInvitingUsersToBoard(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)

	user1, client1, err := cb.createUser()
	a.Nil(err)
	user2, client2, err := cb.createUser()
	a.Nil(err)
	user3, client3, err := cb.createUser()
	a.Nil(err)
	user4, client4, err := cb.createUser()
	a.Nil(err)
	user5, client5, err := cb.createUser()
	a.Nil(err)

	board, resp, err := client1.CreateBoard(client.NewBoard{
		Name:        "This is a test",
		Description: "Just a description, don't mind me.",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	a.Equal("This is a test", board.Name)
	a.Equal("Just a description, don't mind me.", board.Description)
	a.Equal(user1.Uid, board.CreatedBy.UserId)
	boardId := board.BoardId

	// create two normal invites and one specific to user4
	invite, resp, err := client1.CreateInvite(board.BoardId, client.NewInvite{
		Role: client.RoleEditor,
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	a.Equal(board.BoardId, invite.BoardId)
	a.Equal(client.RoleEditor, invite.Role)
	a.Equal(user1.Uid, invite.CreatedBy.UserId)

	invite2, resp, err := client1.CreateInvite(board.BoardId, client.NewInvite{
		Role: client.RoleViewer,
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	a.Equal(board.BoardId, invite2.BoardId)
	a.Equal(client.RoleViewer, invite2.Role)
	a.Equal(user1.Uid, invite2.CreatedBy.UserId)

	invite3, resp, err := client1.CreateInvite(board.BoardId, client.NewInvite{
		User: client.User{
			UserId: user4.Uid,
		},
		Role: client.RoleViewer,
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	a.Equal(board.BoardId, invite3.BoardId)
	a.Equal(client.RoleViewer, invite3.Role)
	a.Equal(user1.Uid, invite3.CreatedBy.UserId)

	board, resp, err = client1.GetBoard(board.BoardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board.Users, 1)
	a.Len(board.Invites, 3)

	resp, err = client2.RespondToInvite(board.BoardId, invite.InviteId, client.InviteResponse{
		Response: client.InviteResponseAccept,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	// user2 has role "editor", so they should be able to see invites and users
	board, resp, err = client2.GetBoard(board.BoardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board.Users, 2)
	a.Len(board.Invites, 2)

	// user3 should not be able to respond to invite3 since it is meant for user4
	resp, err = client3.RespondToInvite(board.BoardId, invite3.InviteId, client.InviteResponse{
		Response: client.InviteResponseAccept,
	})
	a.Nil(err)
	a.Equal(400, resp.HttpCode)
	a.NotZero(resp.ErrorCode)
	a.NotEmpty(resp.ErrorMessage)

	resp, err = client3.RespondToInvite(board.BoardId, invite2.InviteId, client.InviteResponse{
		Response: client.InviteResponseAccept,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	// user3 has role "viewer", they should NOT be able to see invites and users
	board, resp, err = client3.GetBoard(board.BoardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board.Users, 0)
	a.Len(board.Invites, 0)

	resp, err = client4.RespondToInvite(board.BoardId, invite3.InviteId, client.InviteResponse{
		Response: client.InviteResponseAccept,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	// there should be no more invites
	board, resp, err = client1.GetBoard(board.BoardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board.Users, 4)
	a.Len(board.Invites, 0)
	a.ElementsMatch([]string{
		user1.Uid, user2.Uid, user3.Uid, user4.Uid,
	}, userIdsFromUsers(board.Users))

	// user5 is not on the board, can't access it
	newBoard, resp, err := client5.GetBoard(boardId)
	a.Nil(err)
	a.Equal(403, resp.HttpCode)
	a.Empty(newBoard)

	// user2 has role editor, should be able to create an invite
	invite5, resp, err := client2.CreateInvite(board.BoardId, client.NewInvite{
		Role: client.RoleViewer,
		User: client.User{
			UserId: user5.Uid,
		},
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	a.Equal(board.BoardId, invite5.BoardId)
	a.Equal(client.RoleViewer, invite5.Role)
	a.Equal(user2.Uid, invite5.CreatedBy.UserId)
	a.Equal(user5.Uid, invite5.User.UserId)

	invite6, resp, err := client2.CreateInvite(board.BoardId, client.NewInvite{
		Role: client.RoleViewer,
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)

	// shouldn't be able to create an invite for the same user again
	invite7, resp, err := client2.CreateInvite(board.BoardId, client.NewInvite{
		Role: client.RoleViewer,
		User: client.User{
			UserId: user5.Uid,
		},
	})
	a.Nil(err)
	a.Equal(400, resp.HttpCode)
	a.Empty(invite7)
	a.NotZero(resp.ErrorCode)
	a.NotEmpty(resp.ErrorMessage)

	// there should now be 2 invites
	board2, resp, err := client2.GetBoard(board.BoardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board2.Users, 4)
	a.Len(board2.Invites, 2)

	// decline one invite and delete the other
	resp, err = client5.RespondToInvite(boardId, invite5.InviteId, client.InviteResponse{
		Response: client.InviteResponseDecline,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	resp, err = client2.DeleteInvite(boardId, invite6.InviteId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	// there should now be 0 invites
	board2, resp, err = client2.GetBoard(board.BoardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board2.Users, 4)
	a.Len(board2.Invites, 0)
}

func TestEditingBoardUsers(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)

	user1, client1, err := cb.createUser()
	a.Nil(err)
	user2, client2, err := cb.createUser()
	a.Nil(err)
	user3, client3, err := cb.createUser()
	a.Nil(err)

	board, resp, err := client1.CreateBoard(client.NewBoard{
		Name:        "test board 123",
		Description: "abc",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	a.Equal("test board 123", board.Name)
	a.Equal("abc", board.Description)

	boardId := board.BoardId

	// add user2 and user3 to the board
	invite2, resp, err := client1.CreateInvite(boardId, client.NewInvite{
		User: client.User{UserId: user2.Uid},
		Role: client.RoleViewer,
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	invite3, resp, err := client1.CreateInvite(boardId, client.NewInvite{
		User: client.User{UserId: user3.Uid},
		Role: client.RoleViewer,
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	// accept the invites
	resp, err = client2.RespondToInvite(boardId, invite2.InviteId, client.InviteResponse{Response: client.InviteResponseAccept})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	resp, err = client3.RespondToInvite(boardId, invite3.InviteId, client.InviteResponse{Response: client.InviteResponseAccept})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	board, resp, err = client1.GetBoard(boardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board.Users, 3)
	a.ElementsMatch([]string{user1.Uid, user2.Uid, user3.Uid}, userIdsFromUsers(board.Users))

	// user2 has role viewer, should not be able to remove user4
	resp, err = client2.RemoveUserFromBoard(boardId, user3.Uid)
	a.Nil(err)
	a.Equal(403, resp.HttpCode)

	// change the role of user2 to be editor, they should then be able to delete/remove user3 from the board
	boardUser, resp, err := client1.EditBoardUser(boardId, user2.Uid, client.BoardUserEdit{}.WithRole(client.RoleEditor))
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(client.RoleEditor, boardUser.Role)
	a.Equal(user2.Uid, boardUser.User.UserId)
	a.Equal(user1.Uid, boardUser.ModifiedBy.UserId)

	resp, err = client2.RemoveUserFromBoard(boardId, user3.Uid)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	board, resp, err = client2.GetBoard(boardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(board.Invites, 0)
	a.Len(board.Users, 2)
	a.ElementsMatch([]string{user1.Uid, user2.Uid}, userIdsFromUsers(board.Users))

	// client 3 should not longer be able to get board
	board, resp, err = client3.GetBoard(boardId)
	a.Nil(err)
	a.Equal(403, resp.HttpCode)
	a.Empty(board)
}

func TestGetInvitesAndBoardsForUser(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)

	_, client1, err := cb.createUser()
	a.Nil(err)
	user2, client2, err := cb.createUser()
	a.Nil(err)

	// create multiple boards, and send multiple invites
	b1, resp, err := client1.CreateBoard(client.NewBoard{
		Name: "board1",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	b2, resp, err := client1.CreateBoard(client.NewBoard{
		Name: "board2",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	b3, resp, err := client1.CreateBoard(client.NewBoard{
		Name: "board3",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)

	i1, resp, err := client1.CreateInvite(b1.BoardId, client.NewInvite{
		Role: client.RoleViewer,
		User: client.User{
			UserId: user2.Uid,
		},
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	i2, resp, err := client1.CreateInvite(b2.BoardId, client.NewInvite{
		Role: client.RoleViewer,
		User: client.User{
			UserId: user2.Uid,
		},
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	i3, resp, err := client1.CreateInvite(b3.BoardId, client.NewInvite{
		Role: client.RoleViewer,
		User: client.User{
			UserId: user2.Uid,
		},
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)

	invites, resp, err := client2.GetInvites(client.QueryParams{})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.ElementsMatch([]string{i1.InviteId, i2.InviteId, i3.InviteId}, inviteIdsFromInvites(invites))

	boards, resp, err := client1.GetBoards(client.QueryParams{})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.ElementsMatch([]string{b1.BoardId, b2.BoardId, b3.BoardId}, boardIdsFromBoards(boards))
}

func TestDeleteBoard(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)

	_, client1, err := cb.createUser()
	a.Nil(err)

	board, resp, err := client1.CreateBoard(client.NewBoard{
		Name: "board1",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	boardId := board.BoardId

	newBoard, resp, err := client1.GetBoard(boardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(newBoard.BoardId, board.BoardId)
	a.Equal(newBoard.Users, board.Users)

	resp, err = client1.DeleteBoard(boardId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	newBoard, resp, err = client1.GetBoard(boardId)
	a.Nil(err)
	a.Equal(403, resp.HttpCode)
	a.Empty(newBoard)
}

func TestLink(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)

	_, client1, err := cb.createUser()
	a.Nil(err)
	_, client2, err := cb.createUser()
	a.Nil(err)
	_, client3, err := cb.createUser()
	a.Nil(err)
	_, client4, err := cb.createUser()
	a.Nil(err)

	boardId, err := createBoardWithUsers(t, []*client.ApiClient{
		client1, client2, client3, client4,
	})
	a.Nil(err)

	// create link and let other users rate it

	link, resp, err := client2.CreateLink(boardId, client.NewLink{
		Title: "just a link",
		Url:   "https://justarandomhost.com/amazing.png",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	linkId := link.LinkId

	for _, c := range []*client.ApiClient{client1, client2, client3} {
		resp, err := c.RateLink(boardId, linkId, client.LinkRating{
			Rating: 1,
		})
		a.Nil(err)
		a.Equal(200, resp.HttpCode)
	}

	link, resp, err = client1.GetLink(boardId, linkId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(linkId, link.LinkId)
	a.Equal("just a link", link.Title)
	a.Equal("https://justarandomhost.com/amazing.png", link.Url)
	a.Equal(3, link.Score)
	a.Equal(0, link.Downvotes)
	a.Equal(3, link.Upvotes)
	a.Equal(1, link.UserRating)

	// user 4 has not rated yet, should have 0 user rating
	link, resp, err = client4.GetLink(boardId, linkId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(3, link.Score)
	a.Equal(0, link.Downvotes)
	a.Equal(3, link.Upvotes)
	a.Equal(0, link.UserRating)

	// user 4 downvotes
	resp, err = client4.RateLink(boardId, linkId, client.LinkRating{Rating: -1})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	link, resp, err = client4.GetLink(boardId, linkId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Equal(2, link.Score)
	a.Equal(-1, link.Downvotes)
	a.Equal(3, link.Upvotes)
	a.Equal(-1, link.UserRating)

	// delete link, client4 shouldn't be able to, but client 2 and 1 can
	resp, err = client4.DeleteLink(boardId, linkId)
	a.Nil(err)
	a.Equal(403, resp.HttpCode)

	resp, err = client1.DeleteLink(boardId, linkId)
	a.Nil(err)
	a.Equal(200, resp.HttpCode)

	link, resp, err = client2.GetLink(boardId, linkId)
	a.Nil(err)
	a.Equal(404, resp.HttpCode)
	a.Empty(link)
}

type boardQueryElement struct {
	BoardId     string
	CreatedTime int64
}

func boardQueryElementsFromBoards(boards []client.Board) []boardQueryElement {
	result := make([]boardQueryElement, len(boards))
	for i, board := range boards {
		result[i] = boardQueryElement{
			BoardId:     board.BoardId,
			CreatedTime: board.CreatedTime,
		}
	}
	return result
}

func sortBoardQueryElements(boards []boardQueryElement) []boardQueryElement {
	result := make([]boardQueryElement, len(boards))
	copy(result, boards)

	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedTime > result[j].CreatedTime {
			return true
		}
		return false
	})

	return result
}

type inviteQueryElement struct {
	BoardId     string
	InviteId    string
	CreatedTime int64
}

func inviteQueryElementsFromInvites(invites []client.BoardInvite) []inviteQueryElement {
	result := make([]inviteQueryElement, len(invites))
	for i, invite := range invites {
		result[i] = inviteQueryElement{
			BoardId:     invite.BoardId,
			InviteId:    invite.InviteId,
			CreatedTime: invite.CreatedTime,
		}
	}
	return result
}

func sortInviteQueryElements(invites []inviteQueryElement) []inviteQueryElement {
	result := make([]inviteQueryElement, len(invites))
	copy(result, invites)

	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedTime > result[j].CreatedTime {
			return true
		}
		return false
	})

	return result
}

func TestBoardsAndInvitesQueryAndPagination(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)
	_, client1, err := cb.createUser()
	a.Nil(err)
	user2, client2, err := cb.createUser()
	a.Nil(err)

	boardCount := 20
	invites := make([]inviteQueryElement, boardCount)
	boards := make([]boardQueryElement, boardCount)
	for i := 0; i < boardCount; i++ {
		board, resp, err := client1.CreateBoard(client.NewBoard{
			Name: "board1",
		})
		a.Nil(err)
		a.Equal(201, resp.HttpCode)
		boardId := board.BoardId
		boards[i] = boardQueryElement{
			BoardId:     boardId,
			CreatedTime: board.CreatedTime,
		}

		invite, resp, err := client1.CreateInvite(boardId, client.NewInvite{
			User: client.User{UserId: user2.Uid},
			Role: client.RoleViewer,
		})
		a.Nil(err)
		a.Equal(201, resp.HttpCode)
		invites[i] = inviteQueryElement{
			BoardId:     boardId,
			InviteId:    invite.InviteId,
			CreatedTime: invite.CreatedTime,
		}
	}
	sortedBoards := sortBoardQueryElements(boards)
	sortedInvites := sortInviteQueryElements(invites)

	// test board queries and pagination
	br, resp, err := client1.GetBoards(client.QueryParams{
		Limit: 11,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(br, 11)
	a.Equal(sortedBoards[:11], boardQueryElementsFromBoards(br))

	br, resp, err = client1.GetBoards(client.QueryParams{
		Limit:  11,
		Cursor: sortedBoards[10].CreatedTime - 10,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(br, boardCount-11)
	a.Equal(sortedBoards[11:], boardQueryElementsFromBoards(br))

	// test invite queries and pagination
	ir, resp, err := client2.GetInvites(client.QueryParams{
		Limit: 13,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(ir, 13)
	a.Equal(sortedInvites[:13], inviteQueryElementsFromInvites(ir))

	ir, resp, err = client2.GetInvites(client.QueryParams{
		Limit:  13,
		Cursor: sortedInvites[12].CreatedTime - 3,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(ir, boardCount-13)
	a.Equal(sortedInvites[13:], inviteQueryElementsFromInvites(ir))
}

func TestLinksQueryAndPagination(t *testing.T) {
	a := assert.New(t)

	cb, err := newClientBuilder()
	a.Nil(err)

	scores := []int{
		4, 10, 3, 8, 1, 0, 15, 4, 3, 11, 10, 32, 1,
	}
	boardId, linkIds, clients, err := createLinksWithScoes(t, cb, scores)
	a.Nil(err)

	c := clients[0]
	links, resp, err := c.GetLinks(boardId, client.LinkQueryParams{
		Limit: 10,
		Sort:  client.LinkSortTop,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(links, 10)
	linkIdsByScore := sortLinksByScore(linkIds)
	a.Equal(linkIdsByScore[:10], linkIdSortFromLinks(links))

	// load remaining results
	scoreCursor := linkIdsByScore[9].Score
	timeCreatedCursor := linkIdsByScore[9].CreatedTime - 2
	links, resp, err = c.GetLinks(boardId, client.LinkQueryParams{
		Limit:             10,
		Sort:              client.LinkSortTop,
		CursorScore:       &scoreCursor,
		CursorCreatedTime: &timeCreatedCursor,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(links, len(scores)-10)
	a.Equal(linkIdsByScore[10:], linkIdSortFromLinks(links))

	// sort by newest
	links, resp, err = c.GetLinks(boardId, client.LinkQueryParams{
		Limit: 10,
		Sort:  client.LinkSortNewest,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(links, 10)
	linkIdsByNewest := sortLinksByNewest(linkIds)
	a.Equal(linkIdsByNewest[:10], linkIdSortFromLinks(links))

	// load remaining results
	lastCreatedTime := linkIdsByNewest[9].CreatedTime - 2
	links, resp, err = c.GetLinks(boardId, client.LinkQueryParams{
		Limit:             10,
		Sort:              client.LinkSortNewest,
		CursorCreatedTime: &lastCreatedTime,
	})
	a.Nil(err)
	a.Equal(200, resp.HttpCode)
	a.Len(links, len(scores)-10)
	a.Equal(linkIdsByNewest[10:], linkIdSortFromLinks(links))
}

func sortLinksByScore(links []linkIdSort) []linkIdSort {
	result := make([]linkIdSort, len(links))
	for i, link := range links {
		result[i] = link
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Score > result[j].Score {
			return true
		} else if result[i].Score == result[j].Score &&
			result[i].CreatedTime > result[j].CreatedTime {
			return true
		}
		return false
	})
	return result
}

func sortLinksByNewest(links []linkIdSort) []linkIdSort {
	result := make([]linkIdSort, len(links))
	for i, link := range links {
		result[i] = link
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedTime > result[j].CreatedTime {
			return true
		}
		return false
	})
	return result
}

func linkIdSortFromLinks(links []client.Link) []linkIdSort {
	result := make([]linkIdSort, len(links))
	for i, link := range links {
		result[i] = linkIdSort{
			LinkId:      link.LinkId,
			CreatedTime: link.CreatedTime,
			Score:       link.Score,
		}
	}
	return result
}

// first client passed will create the board, others will be added with viewer role
func createBoardWithUsers(t *testing.T, clients []*client.ApiClient) (string, error) {
	a := assert.New(t)

	if len(clients) < 1 {
		return "", errors.New("need to pass at least one client")
	}
	client1 := clients[0]

	board, resp, err := client1.CreateBoard(client.NewBoard{
		Name: "board1",
	})
	a.Nil(err)
	a.Equal(201, resp.HttpCode)
	boardId := board.BoardId

	for _, c := range clients[1:] {
		invite, resp, err := client1.CreateInvite(boardId, client.NewInvite{
			Role: client.RoleViewer,
		})
		a.Nil(err)
		a.Equal(201, resp.HttpCode)
		resp, err = c.RespondToInvite(boardId, invite.InviteId, client.InviteResponse{
			Response: client.InviteResponseAccept,
		})
		a.Nil(err)
		a.Equal(200, resp.HttpCode)
	}

	return boardId, nil
}

type linkIdSort struct {
	LinkId      string
	CreatedTime int64
	Score       int
}

func createLinksWithScoes(t *testing.T, cb *clientBuilder, scores []int) (string, []linkIdSort, []*client.ApiClient, error) {
	a := assert.New(t)

	if len(scores) == 0 {
		return "", nil, nil, errors.New("empty list of scores")
	}

	clientsNeeded := 1
	for _, score := range scores {
		absScore := score
		if score < 0 {
			absScore = score * -1
		}
		if absScore > clientsNeeded {
			clientsNeeded = absScore
		}
	}

	// cannot have link with score > 32, because number of users per board is limited
	if clientsNeeded > 32 {
		return "", nil, nil, fmt.Errorf("cannot have link with score: %v", clientsNeeded)
	}

	clients := make([]*client.ApiClient, clientsNeeded)
	for i := 0; i < clientsNeeded; i++ {
		_, client, err := cb.createUser()
		if err != nil {
			return "", nil, nil, err
		}
		clients[i] = client
	}

	boardId, err := createBoardWithUsers(t, clients)
	if err != nil {
		return "", nil, nil, err
	}

	linkCreater := clients[0]

	// create links and vote for them
	linkIds := make([]linkIdSort, len(scores))
	for i, score := range scores {
		link, resp, err := linkCreater.CreateLink(boardId, client.NewLink{
			Title: "some title",
			Url:   "https://justaurl.com/amazing.png",
		})
		a.Nil(err)
		a.Equal(201, resp.HttpCode)
		linkIds[i] = linkIdSort{
			LinkId:      link.LinkId,
			CreatedTime: link.CreatedTime,
			Score:       score,
		}

		absScore := score
		rating := 1
		if score < 0 {
			absScore = score * -1
			rating = -1
		}

		for j := 0; j < absScore; j++ {
			c := clients[j]
			resp, err := c.RateLink(boardId, link.LinkId, client.LinkRating{
				Rating: rating,
			})
			a.Nil(err)
			a.Equal(200, resp.HttpCode)
		}
	}

	return boardId, linkIds, clients, nil
}

func userIdsFromUsers(users []client.BoardUser) []string {
	result := make([]string, len(users))
	for i, user := range users {
		result[i] = user.User.UserId
	}
	return result
}

func inviteIdsFromInvites(invites []client.BoardInvite) []string {
	result := make([]string, len(invites))
	for i, invite := range invites {
		result[i] = invite.InviteId
	}
	return result
}

func boardIdsFromBoards(boards []client.Board) []string {
	result := make([]string, len(boards))
	for i, board := range boards {
		result[i] = board.BoardId
	}
	return result
}

type user struct {
	Uid   string
	Email string
	Token string
}

type clientBuilder struct {
	authClient *emulator.AuthEmulatorClient
	address    string
}

// create a new user with a random email address
func (u clientBuilder) createUser() (user, *client.ApiClient, error) {
	email := fmt.Sprintf("%v@test.de", rand.Int63())

	uid, err := u.authClient.CreateUser(email, "test123", true)
	if err != nil {
		return user{}, nil, err
	}
	token, err := u.authClient.SignInUser(email, "test123")
	if err != nil {
		return user{}, nil, err
	}

	apiClient := client.NewApiClient(uid, token, u.address)
	return user{
		Uid:   uid,
		Email: email,
		Token: token,
	}, apiClient, nil
}
