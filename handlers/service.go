package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/maershov/tp_db_forum.git/queries"
)

func ClearDatabase(w http.ResponseWriter, r *http.Request) {
	err := queries.ClearDatabase()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetDatabaseStatus(w http.ResponseWriter, r *http.Request) {
	res, err := queries.GetDatabaseStatus()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j, err := res.MarshalJSON()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}
