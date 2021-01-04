package queries

import (
	"database/sql"
	"strings"

	"github.com/lib/pq"

	"github.com/maershov/tp_db_forum.git/models"
)

func CreateForum(f *models.Forum) (*models.Forum, error) {
	if f.ForumTitle == "" || f.ForumSlug == "" || f.ForumUser == "" {
		return nil, &NullFieldError{"Forum", "title and/or slug and/or user"}
	}

	res := &models.Forum{}
	err := db.Get(
		res,
		`INSERT INTO forum (forum_title, forum_slug, forum_user)
		VALUES ($1, $2, (SELECT nickname FROM forum_user WHERE nickname = $3)) RETURNING *`,
		f.ForumTitle, f.ForumSlug, f.ForumUser)
	if err != nil {
		pqErr := err.(*pq.Error)
		switch pqErr.Code {
		case UniqueViolationCode:
			if strings.HasPrefix(pqErr.Detail, "Key (forum_slug)") {
				res, err := GetForumBySlug(f.ForumSlug)
				if err != nil {
					return res, err
				}
				return res, &UniqueFieldValueAlreadyExistsError{"Forum", "slug"}
			}
		case NotNullViolationCode:
			if pqErr.Column == "forum_user" {
				return res, &RecordNotFoundError{"User", f.ForumUser}
			}
		}
		return res, err
	}

	return res, nil
}

func GetForumBySlug(s string) (*models.Forum, error) {
	res := &models.Forum{}
	err := db.Get(res, "SELECT * FROM forum	WHERE forum_slug = $1", s)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, &RecordNotFoundError{"Forum", s}
		}
		return res, err
	}
	return res, nil
}

func CheckExistenceOfForum(s string) error {
	err := db.QueryRow("SELECT FROM forum WHERE forum_slug = $1", s).Scan()
	if err != nil {
		if err == sql.ErrNoRows {
			return &RecordNotFoundError{"Forum", s}
		}
		return err
	}

	return nil
}
