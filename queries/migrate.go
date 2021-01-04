package queries

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/rubenv/sql-migrate" // applies migrations
)

func makeMigrations(db *sqlx.DB) {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Println(err)
	} else if n != 0 {
		log.Printf("Applied %d migrations!\n", n)
	}
}
