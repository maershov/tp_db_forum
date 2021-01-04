package queries

import (
	"github.com/lib/pq"

	"github.com/maershov/tp_db_forum.git/models"
)

func VoteForPost(v *models.Vote, path string) (*models.Thread, error) {
	if v.Nickname == "" {
		return nil, &NullFieldError{"Vote", "nickname"}
	}
	if v.Voice != -1 && v.Voice != 1 {
		return nil, &ValidationError{"Vote", "voice"}
	}

	threadID, err := GetThreadIDBySlugOrID(path)
	if err != nil {
		return nil, err
	}

	res := &models.Thread{}
	_, err = db.Exec(`
		INSERT INTO vote VALUES (
			(SELECT nickname FROM forum_user WHERE nickname = $1), $2, $3
		)
		ON CONFLICT (nickname, thread) DO UPDATE SET voice = $3`,
		v.Nickname, threadID, v.Voice)
	if err != nil {
		if pqErr := err.(*pq.Error); pqErr.Code == NotNullViolationCode && pqErr.Column == "nickname" {
			return res, &RecordNotFoundError{"User", v.Nickname}
		}
		return res, err
	}

	res, err = GetThreadByID(threadID)
	if err != nil {
		return res, err
	}
	return res, nil
}
