package queries

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/maershov/tp_db_forum.git/models"
)

func CreatePosts(p *models.PostList, path string) (*models.PostList, error) {
	t, err := GetThreadBySlugOrID(path)
	if err != nil {
		return nil, err
	}

	if len(*p) == 0 {
		return &models.PostList{}, nil
	}

	// get current time, we'll use it for all inserted messages
	now := time.Time{}
	err = db.QueryRow("SELECT * FROM now()").Scan(&now)
	if err != nil {
		return nil, err
	}

	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ustmt, err := tx.Prepare("SELECT nickname FROM forum_user WHERE nickname = $1")
	if err != nil {
		return nil, err
	}
	defer ustmt.Close()
	idstmt, err := tx.Prepare("SELECT nextval(pg_get_serial_sequence('post', 'post_id'))")
	if err != nil {
		return nil, err
	}
	defer idstmt.Close()
	poststmt, err := tx.Preparex("SELECT * FROM post WHERE post_id = $1")
	if err != nil {
		return nil, err
	}
	defer poststmt.Close()
	for k, v := range *p {
		// get user
		err := ustmt.QueryRow(v.PostAuthor).Scan(&(*p)[k].PostAuthor)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, &RecordNotFoundError{"User", v.PostAuthor}
			}
			return nil, err
		}

		// check parent message belongs to the same thread
		if v.Parent != 0 {
			parent := models.Post{}
			err = poststmt.QueryRow(v.Parent).Scan(
				&parent.PostID, &parent.Forum, &parent.Thread, &parent.Parent, pq.Array(&parent.Path),
				&parent.Path1, &parent.PostAuthor, &parent.PostCreated, &parent.IsEdited, &parent.PostMessage)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil, ErrParentPostIsNotInThisThread // not exists
				}
				return nil, err
			}
			if parent.Thread != t.ThreadID {
				return nil, ErrParentPostIsNotInThisThread
			}
			(*p)[k].Path = parent.Path
		}

		// get new primary key id
		err = idstmt.QueryRow().Scan(&(*p)[k].PostID)
		if err != nil {
			return nil, err
		}
		(*p)[k].Path = append((*p)[k].Path, int64((*p)[k].PostID))

		// update result
		(*p)[k].Forum = t.Forum
		(*p)[k].Thread = t.ThreadID
		(*p)[k].PostCreated = now
	}

	stmt, err := tx.Prepare(pq.CopyIn("post", "post_id", "forum", "thread",
		"parent", "path", "path1", "post_author", "post_created", "post_message"))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for _, v := range *p {
		p1 := v.PostID
		if v.Parent != 0 {
			p1 = int(v.Path[0])
		}
		_, err = stmt.Exec(v.PostID, t.Forum, t.ThreadID,
			v.Parent, pq.Array(v.Path), p1, v.PostAuthor, now, v.PostMessage)
		if err != nil {
			return nil, err
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec("UPDATE forum SET posts = posts + $1 WHERE forum_slug = $2", len(*p), t.Forum)
	if err != nil {
		return nil, err
	}

	res := &models.PostList{}
	*res = *p

	err = tx.Commit()
	if err != nil {
		return res, err
	}

	// insert without transaction!
	uifstmt, err := db.Prepare(`
		INSERT INTO users_in_forum (forum_user, forum) VALUES ($1, $2) 
		ON CONFLICT (forum_user, forum) DO NOTHING`)
	if err != nil {
		return res, err
	}
	defer uifstmt.Close()
	for _, v := range *p {
		_, err = uifstmt.Exec(
			v.PostAuthor, t.Forum)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}

func GetPostByID(id int) (*models.Post, error) {
	res := &models.Post{}
	err := db.QueryRow("SELECT * FROM post WHERE post_id = $1", id).Scan(
		&res.PostID, &res.Forum, &res.Thread, &res.Parent, pq.Array(&res.Path), &res.Path1, &res.PostAuthor,
		&res.PostCreated, &res.IsEdited, &res.PostMessage)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return res, err
	}

	return res, nil
}

func GetPostInfoByID(id int, params *[]string) (*models.PostInfo, error) {
	q := strings.Builder{}
	q.WriteString("SELECT post_id, p.forum forum_slug, thread, parent, post_author, post_created, is_edited, post_message")
	queryArgs := make(map[string]bool, 3)
	for _, v := range *params {
		queryArgs[v] = true
	}
	if _, ok := queryArgs["user"]; ok {
		q.WriteString(", u.about, u.email, u.fullname")
	}
	if _, ok := queryArgs["thread"]; ok {
		q.WriteString(", thread_id, thread_slug, thread_title, thread_author, thread_created, thread_message, votes")
	}
	if _, ok := queryArgs["forum"]; ok {
		q.WriteString(", forum_title, forum_user, threads, posts")
	}

	q.WriteString(" FROM post p")
	if _, ok := queryArgs["user"]; ok {
		q.WriteString(" JOIN forum_user u ON post_author = u.nickname")
	}
	if _, ok := queryArgs["thread"]; ok {
		q.WriteString(" JOIN thread t ON p.thread = t.thread_id")
	}
	if _, ok := queryArgs["forum"]; ok {
		q.WriteString(" JOIN forum f ON p.forum = f.forum_slug")
	}
	q.WriteString(" WHERE post_id = $1")

	all := &models.PostInfoAllFields{}
	err := db.QueryRowx(q.String(), id).StructScan(all)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return nil, err
	}

	res := &models.PostInfo{}
	res.Post = &models.Post{
		PostID:      all.PostID,
		Forum:       all.ForumSlug,
		Thread:      all.Post.Thread,
		Parent:      all.Parent,
		PostAuthor:  all.PostAuthor,
		PostCreated: all.PostCreated,
		IsEdited:    all.IsEdited,
		PostMessage: all.PostMessage,
	}
	if _, ok := queryArgs["user"]; ok {
		res.Author = &models.ForumUser{
			Nickname: all.PostAuthor,
			Fullname: all.Fullname,
			Email:    all.Email,
			About:    all.About,
		}
	}
	if _, ok := queryArgs["thread"]; ok {
		res.Thread = &models.Thread{
			ThreadID:      all.ThreadID,
			Forum:         all.ForumSlug,
			ThreadSlug:    all.ThreadSlug,
			ThreadTitle:   all.ThreadTitle,
			ThreadAuthor:  all.ThreadAuthor,
			ThreadCreated: all.ThreadCreated,
			ThreadMessage: all.ThreadMessage,
			Votes:         all.Votes,
		}
	}
	if _, ok := queryArgs["forum"]; ok {
		res.Forum = &models.Forum{
			ForumTitle: all.ForumTitle,
			ForumSlug:  all.ForumSlug,
			ForumUser:  all.Forum.ForumUser,
			Threads:    all.Threads,
			Posts:      all.Posts,
		}
	}

	return res, nil
}

func UpdatePostByID(id int, p *models.Post) (*models.Post, error) {
	if p.PostMessage == "" {
		return GetPostByID(id)
	}
	res := &models.Post{}
	err := db.Get(res,
		`UPDATE post SET 
			post_message = $1, 
			is_edited = CASE WHEN $1 <> (SELECT post_message FROM post WHERE post_id = $2) 
				THEN TRUE 
				ELSE FALSE
			END 
		WHERE post_id = $2
		RETURNING post_id, forum, thread, parent, post_author, post_created, is_edited, post_message`,
		p.PostMessage, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Post", fmt.Sprintf("%v", id)}
		}
		return res, err
	}

	return res, nil
}

func GetThreadPosts(slugOrID string, args *models.ThreadPostsQueryArgs) (*models.PostList, error) {
	threadID, err := GetThreadIDBySlugOrID(slugOrID)
	if err != nil {
		return nil, err
	}
	q := strings.Builder{}
	q.WriteString(`SELECT p.post_id, p.forum, p.thread, p.parent, 
		p.post_author, p.post_created, p.is_edited, p.post_message FROM post p `)
	switch args.Sort {
	case "tree":
		if args.Since > 0 {
			q.WriteString("JOIN post sp ON sp.post_id = $2 WHERE p.path ")
			if args.Desc {
				q.WriteString("< sp.path ")
			} else {
				q.WriteString("> sp.path ")
			}
			q.WriteString("AND p.thread = $1")
		} else {
			q.WriteString("WHERE p.thread = $1")
		}
		q.WriteString(" ORDER BY p.path")
		if args.Desc {
			q.WriteString(" DESC")
		}
		if args.Limit > 0 {
			if args.Since > 0 {
				q.WriteString(" LIMIT $3")
			} else {
				q.WriteString(" LIMIT $2")
			}
		}
	case "parent_tree":
		q.WriteString("WHERE p.path1 IN (SELECT p.post_id FROM post p ")
		if args.Since > 0 {
			q.WriteString("JOIN post sp ON sp.post_id = $2 ")
			if args.Desc {
				q.WriteString("WHERE p.path1 < sp.path1 AND p.thread = $1 AND p.parent = 0")
			} else {
				q.WriteString("WHERE p.path > sp.path AND p.thread = $1 AND p.parent = 0")
			}
		} else {
			q.WriteString("WHERE p.thread = $1 AND p.parent = 0 ")
		}
		q.WriteString("ORDER BY p.path1")
		if args.Desc {
			q.WriteString(" DESC")
		}
		if args.Limit > 0 {
			if args.Since > 0 {
				q.WriteString(" LIMIT $3")
			} else {
				q.WriteString(" LIMIT $2")
			}
		}
		q.WriteString(`) ORDER BY p.path1`)
		if args.Desc {
			q.WriteString(" DESC")
		}
		q.WriteString(", p.path")
	default: // flat
		q.WriteString(" WHERE p.thread = $1")
		if args.Since > 0 {
			if args.Desc {
				q.WriteString(" AND p.post_id < $2")
			} else {
				q.WriteString(" AND p.post_id > $2")
			}
		}
		q.WriteString(" ORDER BY p.post_created")
		if args.Desc {
			q.WriteString(" DESC")
		}
		q.WriteString(", p.post_id")
		if args.Desc {
			q.WriteString(" DESC")
		}
		if args.Limit > 0 {
			if args.Since > 0 {
				q.WriteString(" LIMIT $3")
			} else {
				q.WriteString(" LIMIT $2")
			}
		}
	}
	res := &models.PostList{}
	if args.Since > 0 {
		if args.Limit > 0 {
			err = db.Select(res, q.String(), threadID, args.Since, args.Limit)
		} else {
			err = db.Select(res, q.String(), threadID, args.Since)
		}

	} else {
		if args.Limit > 0 {
			err = db.Select(res, q.String(), threadID, args.Limit)
		} else {
			err = db.Select(res, q.String(), threadID)
		}
	}
	if err != nil {
		return res, err
	}

	return res, nil
}
