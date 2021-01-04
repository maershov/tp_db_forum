package models

//easyjson:json
type ForumUser struct {
	Nickname string `json:"nickname"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

//easyjson:json
type ForumUserList []ForumUser
