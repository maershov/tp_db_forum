package queries

import (
	"github.com/maershov/tp_db_forum.git/models"
)

func ClearDatabase() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("TRUNCATE TABLE users_in_forum CASCADE")
	if err != nil {
		return err
	}
	_, err = tx.Exec("TRUNCATE TABLE vote CASCADE")
	if err != nil {
		return err
	}
	_, err = tx.Exec("TRUNCATE TABLE post CASCADE")
	if err != nil {
		return err
	}
	_, err = tx.Exec("TRUNCATE TABLE thread CASCADE")
	if err != nil {
		return err
	}
	_, err = tx.Exec("TRUNCATE TABLE forum CASCADE")
	if err != nil {
		return err
	}
	_, err = tx.Exec("TRUNCATE TABLE forum_user CASCADE")
	if err != nil {
		return err
	}
	return tx.Commit()
}

func GetDatabaseStatus() (*models.Status, error) {
	res := &models.Status{}
	err := db.Get(res, `SELECT "user", forum, thread, post
		FROM (SELECT COUNT(*) AS "user" FROM forum_user) a
		CROSS JOIN (SELECT COUNT(*) AS forum FROM forum) b
		CROSS JOIN (SELECT COUNT(*) AS thread FROM thread) c
		CROSS JOIN (SELECT COUNT(*) AS post FROM post) d;`)
	if err != nil {
		return res, err
	}
	return res, nil
}
