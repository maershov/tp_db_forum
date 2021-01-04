package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/maershov/tp_db_forum.git/models"
	"github.com/maershov/tp_db_forum.git/queries"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	u := &models.ForumUser{}
	err = u.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	u.Nickname = mux.Vars(r)["nickname"]

	res, err := queries.CreateUser(u)
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
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// if record has been inserted successfully
	j, err := (*res)[0].MarshalJSON()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, string(j))
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	res, err := queries.GetUserByNickname(mux.Vars(r)["nickname"])
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

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	u := &models.ForumUser{}
	err = u.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := queries.UpdateUser(mux.Vars(r)["nickname"], u)
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
		case *queries.UniqueFieldValueAlreadyExistsError:
			j, jErr := models.ErrorMessage{Message: err.Error()}.MarshalJSON()
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintln(w, string(j))
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
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

func GetForumUsers(w http.ResponseWriter, r *http.Request) {
	params := &models.UserQueryParams{}
	query := r.URL.Query()
	rawDesc := query.Get("desc")
	var err error
	if rawDesc != "" {
		params.Desc, err = strconv.ParseBool(rawDesc)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	rawLimit := query.Get("limit")
	if rawLimit != "" {
		params.Limit, err = strconv.ParseUint(rawLimit, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	params.Since = query.Get("since")
	res, err := queries.GetAllUsersInForum(mux.Vars(r)["slug"], params)
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
