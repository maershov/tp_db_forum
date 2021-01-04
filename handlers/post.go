package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/maershov/tp_db_forum.git/models"
	"github.com/maershov/tp_db_forum.git/queries"
)

func CreatePosts(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	p := &models.PostList{}
	err = p.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := mux.Vars(r)["slug_or_id"]

	res, err := queries.CreatePosts(p, path)
	if err != nil {
		if err == queries.ErrParentPostIsNotInThisThread {
			j, jErr := models.ErrorMessage{Message: err.Error()}.MarshalJSON()
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintln(w, string(j))
			return
		}
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
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, string(j))
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	ids := mux.Vars(r)["id"]
	id, err := strconv.Atoi(ids)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var params []string
	if qs, ok := r.URL.Query()["related"]; ok {
		params = strings.Split(qs[0], ",")
	}

	res, err := queries.GetPostInfoByID(id, &params)
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

	j, err := res.MarshalJSON()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	p := &models.Post{}
	err = p.UnmarshalJSON(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ids := mux.Vars(r)["id"]
	id, err := strconv.Atoi(ids)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := queries.UpdatePostByID(id, p)
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

	j, err := res.MarshalJSON()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(j))
}

func GetThreadPosts(w http.ResponseWriter, r *http.Request) {
	params := &models.ThreadPostsQueryArgs{}
	query := r.URL.Query()
	rawLimit := query.Get("limit")
	var err error
	if rawLimit != "" {
		params.Limit, err = strconv.ParseUint(rawLimit, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	rawSince := query.Get("since")
	if rawSince != "" {
		params.Since, err = strconv.ParseUint(rawSince, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	params.Sort = query.Get("sort")
	rawDesc := query.Get("desc")
	if rawDesc != "" {
		params.Desc, err = strconv.ParseBool(rawDesc)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	path := mux.Vars(r)["slug_or_id"]

	res, err := queries.GetThreadPosts(path, params)
	if err != nil {
		if err == queries.ErrParentPostIsNotInThisThread {
			j, jErr := models.ErrorMessage{Message: err.Error()}.MarshalJSON()
			if jErr != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintln(w, string(j))
			return
		}
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
