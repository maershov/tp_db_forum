package queries

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/maershov/tp_db_forum.git/models"
)

func CreateThread(t *models.Thread) (*models.Thread, error) {
	if t.Forum == "" || t.ThreadTitle == "" || t.ThreadAuthor == "" {
		return nil, &NullFieldError{"Thread", "some value(-s) is/are null"}
	}

	res := &models.Thread{}
	err := db.Get(res, `
		INSERT INTO thread (forum, thread_slug, thread_title, thread_author, thread_created, thread_message)
		VALUES (
			(SELECT forum_slug FROM forum WHERE forum_slug = $1), $2, $3, 
			(SELECT nickname FROM forum_user WHERE nickname = $4), $5, $6
		) RETURNING *`,
		t.Forum, t.ThreadSlug, t.ThreadTitle, t.ThreadAuthor, t.ThreadCreated, t.ThreadMessage)
	if err != nil {
		pqErr := err.(*pq.Error)
		switch pqErr.Code {
		case UniqueViolationCode:
			if strings.HasPrefix(pqErr.Detail, "Key (thread_slug)") {
				res, err := GetThreadBySlug(*t.ThreadSlug)
				if err != nil {
					return res, err
				}
				return res, &UniqueFieldValueAlreadyExistsError{"Thread", "slug"}
			}
		case NotNullViolationCode:
			if pqErr.Column == "thread_author" {
				return res, &RecordNotFoundError{"User", t.ThreadAuthor}
			}
			if pqErr.Column == "forum" {
				return res, &RecordNotFoundError{"Forum", t.Forum}
			}
		}
		return res, err
	}

	_, err = db.Exec(`INSERT INTO users_in_forum (forum_user, forum) 
		VALUES (
			(SELECT nickname FROM forum_user WHERE nickname = $1), 
			(SELECT forum_slug FROM forum WHERE forum_slug = $2)
		) ON CONFLICT (forum_user, forum) DO NOTHING`, t.ThreadAuthor, t.Forum)
	if err != nil {
		return res, err
	}

	return res, nil
}

func GetThreadByID(id int) (*models.Thread, error) {
	res := &models.Thread{}
	err := db.Get(res, "SELECT * FROM thread WHERE thread_id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", fmt.Sprintf("%v", id)}
		}
		return res, err
	}
	return res, nil
}

func GetThreadBySlug(s string) (*models.Thread, error) {
	res := &models.Thread{}
	err := db.Get(res, "SELECT * FROM thread WHERE thread_slug = $1", s)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", s}
		}
		return res, err
	}
	return res, nil
}

func GetThreadBySlugOrID(slugOrID string) (*models.Thread, error) {
	res, err := GetThreadBySlug(slugOrID)
	if err != nil {
		if _, ok := err.(*RecordNotFoundError); ok {
			id, convErr := strconv.Atoi(slugOrID)
			if convErr != nil {
				return res, err
			}
			res, err = GetThreadByID(id)
			if err != nil {
				return res, err
			}
		} else {
			return res, err
		}
	}
	return res, nil
}

func GetThreadIDByID(id int) (int, error) {
	res := 0
	err := db.Get(&res, "SELECT thread_id FROM thread WHERE thread_id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", fmt.Sprintf("%v", id)}
		}
		return res, err
	}
	return res, nil
}

func GetThreadIDBySlug(s string) (int, error) {
	res := 0
	err := db.Get(&res, "SELECT thread_id FROM thread WHERE thread_slug = $1", s)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", s}
		}
		return res, err
	}
	return res, nil
}

func GetThreadIDBySlugOrID(slugOrID string) (int, error) {
	res, err := GetThreadIDBySlug(slugOrID)
	if err != nil {
		if _, ok := err.(*RecordNotFoundError); ok {
			id, convErr := strconv.Atoi(slugOrID)
			if convErr != nil {
				return res, err
			}
			res, err = GetThreadIDByID(id)
			if err != nil {
				return res, err
			}
		} else {
			return res, err
		}
	}
	return res, nil
}

func GetAllThreadsInForum(s string, params *models.ThreadQueryParams) (*models.ThreadList, error) {
	err := CheckExistenceOfForum(s)
	if err != nil {
		return nil, err
	}

	q := strings.Builder{}
	q.WriteString("SELECT * FROM thread WHERE forum = $1 ")
	var nt time.Time
	if params.Since != nt {
		if params.Desc {
			q.WriteString("AND thread_created <= $2\n")
		} else {
			q.WriteString("AND thread_created >= $2\n")
		}
	}
	q.WriteString("ORDER BY thread_created ")
	if params.Desc {
		q.WriteString("DESC")
	}
	if params.Limit != 0 {
		q.WriteString(fmt.Sprintf("\nLIMIT %v", params.Limit))
	}
	res := &models.ThreadList{}
	if params.Since == nt {
		err = db.Select(res, q.String(), s)
	} else {
		err = db.Select(res, q.String(), s, params.Since)
	}
	if err != nil {
		return res, err
	}
	return res, nil
}

func UpdateThread(t *models.Thread, path string) (*models.Thread, error) {
	res, err := GetThreadBySlugOrID(path)
	if err != nil {
		return res, err
	}

	q := strings.Builder{}
	q.WriteString("UPDATE thread SET ")
	args := make([]interface{}, 0, 5)
	continues := false
	fieldCount := 0
	if t.ThreadTitle != "" {
		fieldCount++
		q.WriteString("thread_title = $" + strconv.Itoa(fieldCount))
		continues = true
		args = append(args, t.ThreadTitle)
	}
	if t.ThreadMessage != "" {
		fieldCount++
		if continues {
			q.WriteString(", thread_message = $" + strconv.Itoa(fieldCount))
		} else {
			q.WriteString("thread_message = $" + strconv.Itoa(fieldCount))

		}
		args = append(args, t.ThreadMessage)
	}
	q.WriteString(" WHERE thread_id = $" + strconv.Itoa(fieldCount+1) + " RETURNING *")
	args = append(args, res.ThreadID)
	err = db.Get(res, q.String(), args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Thread", strconv.Itoa(res.ThreadID)}
		}
	}

	return res, nil
}
