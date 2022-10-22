package client

type NewBoard struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type User struct {
	UserId string `json:"userId"`
	Name   string `json:"name"`
}

type Board struct {
	BoardId      string `json:"boardId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	CreatedTime  int64  `json:"createdTime"`
	CreatedBy    User   `json:"createdBy"`
	ModifiedTime int64  `json:"modifiedTime"`
	ModifiedBy   User   `json:"modifiedBy"`

	Users   []BoardUser   `json:"users"`
	Invites []BoardInvite `json:"invites"`
}

type BoardEdit map[string]interface{}

func (b BoardEdit) WithName(name string) BoardEdit {
	b["name"] = name
	return b
}

func (b BoardEdit) WithDescription(description string) BoardEdit {
	b["description"] = description
	return b
}

const RoleOwner = "owner"
const RoleEditor = "editor"
const RoleViewer = "viewer"

type BoardUser struct {
	User         User   `json:"user"`
	Role         string `json:"role"`
	CreatedTime  int64  `json:"createdTime"`
	InvitedBy    User   `json:"invitedBy"`
	Modifiedtime int64  `json:"modifiedTime"`
	ModifiedBy   User   `json:"modifiedBy"`
}

type BoardUserEdit map[string]interface{}

func (b BoardUserEdit) WithRole(role string) BoardUserEdit {
	b["role"] = role
	return b
}

type NewInvite struct {
	User User   `json:"user,omitempty"`
	Role string `json:"role,omitempty"`
}

type BoardInvite struct {
	BoardId  string `json:"boardId"`
	InviteId string `json:"inviteId"`

	User User   `json:"user"`
	Role string `json:"role"`

	CreatedTime int64 `json:"createdTime"`
	CreatedBy   User  `json:"createdBy"`

	ExpiresTime int64 `json:"expiresTime"`
}

const InviteResponseAccept = "accept"
const InviteResponseDecline = "decline"

type InviteResponse struct {
	Response string `json:"response,omitempty"`
}

type QueryParams struct {
	Limit  int
	Cursor int64
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

	CreatedTime int64 `json:"createdTime"`
	CreatedBy   User  `json:"createdBy"`

	Score     int `json:"score"`
	Upvotes   int `json:"upvotes"`
	Downvotes int `json:"downvotes"`

	UserRating int `json:"userRating"`
}

type LinkRating struct {
	Rating int `json:"rating"`
}

const LinkSortNewest = "newest"
const LinkSortTop = "top"

type LinkQueryParams struct {
	Limit             int
	Sort              string
	CursorScore       *int
	CursorCreatedTime *int64
}
