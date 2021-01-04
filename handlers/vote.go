package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/maershov/tp_db_forum.git/models"
	"github.com/maershov/tp_db_forum.git/queries"
)

func VoteForPost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	v := &models.Vote{}
	err = v.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := mux.Vars(r)["slug_or_id"]

	res, err := queries.VoteForPost(v, path)
	if err != nil {
		switch err.(type) {
		case *queries.NullFieldError, *queries.ValidationError:
			j, jErr := models.ErrorMessage{Message: err.Error()}.MarshalJSON()
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, string(j))
		case *queries.RecordNotFoundError:
			j, jErr := models.ErrorMessage{Message: err.Error()}.MarshalJSON()
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, string(j))
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// if records have been inserted successfully
	j, err := res.MarshalJSON()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}
