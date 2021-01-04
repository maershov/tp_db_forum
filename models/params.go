package models

import (
	"time"
)

type ThreadQueryParams struct {
	Desc  bool
	Limit uint64
	Since time.Time
}

type UserQueryParams struct {
	Desc  bool
	Limit uint64
	Since string
}

type PostQueryArgs struct {
	Related string
}

type ThreadPostsQueryArgs struct {
	Limit uint64
	Since uint64
	Sort  string
	Desc  bool
}
