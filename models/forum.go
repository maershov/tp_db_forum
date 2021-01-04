package models

//easyjson:json
type Forum struct {
	ForumSlug  string `json:"slug" db:"forum_slug"`
	ForumTitle string `json:"title" db:"forum_title"`
	ForumUser  string `json:"user" db:"forum_user"`
	Threads    int    `json:"threads"`
	Posts      int    `json:"posts"`
}
