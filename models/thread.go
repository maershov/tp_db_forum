package models

import (
	"time"
)

//easyjson:json
type Thread struct {
	ThreadID      int        `json:"id" db:"thread_id"`
	Forum         string     `json:"forum"`
	ThreadSlug    *string    `json:"slug,omitempty" db:"thread_slug"`
	ThreadTitle   string     `json:"title" db:"thread_title"`
	ThreadAuthor  string     `json:"author" db:"thread_author"`
	ThreadCreated *time.Time `json:"created,omitempty" db:"thread_created"`
	ThreadMessage string     `json:"message" db:"thread_message"`
	Votes         int        `json:"votes"`
}

//easyjson:json
type ThreadList []Thread
