package queries

import (
	"log"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq" // postgres driver
)

var db *sqlx.DB

func InitDB(address, database string) *sqlx.DB {
	var err error
	db, err = sqlx.Open("postgres",
		"postgres://"+address+"/"+database+"?sslmode=disable")
	if err != nil {
		log.Panic(err)
	}

	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	log.Printf("Successfully connected to %v, database %v\n", address, database)

	makeMigrations(db)

	return db
}
