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

func CreateForum(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	f := &models.Forum{}
	err = f.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := queries.CreateForum(f)
	if err != nil {
		switch err.(type) {
		case *queries.NullFieldError:
			j, jErr := models.ErrorMessage{Message: err.Error()}.MarshalJSON()
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, string(j))
		case *queries.UniqueFieldValueAlreadyExistsError:
			j, err := res.MarshalJSON()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
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

	// if record has been inserted successfully
	j, err := res.MarshalJSON()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, string(j))
}

func GetForum(w http.ResponseWriter, r *http.Request) {
	res, err := queries.GetForumBySlug(mux.Vars(r)["slug"])
	if err != nil {
		switch err.(type) {
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

	j, jErr := res.MarshalJSON()
	if jErr != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}
