package main

import (
	"flag"
	"github.com/maershov/tp_db_forum.git/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/maershov/tp_db_forum.git/queries"
)

func main() {
	flag.Parse()

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.Use(handlers.ApplicationJSONMiddleware)

	api.HandleFunc("/forum/create", handlers.CreateForum).Methods("POST")
	api.HandleFunc("/forum/{slug}/create", handlers.CreateThread).Methods("POST")
	api.HandleFunc("/forum/{slug}/details", handlers.GetForum).Methods("GET")
	api.HandleFunc("/forum/{slug}/threads", handlers.GetThreads).Methods("GET")
	api.HandleFunc("/forum/{slug}/users", handlers.GetForumUsers).Methods("GET")

	api.HandleFunc("/post/{id:[0-9]+}/details", handlers.GetPost).Methods("GET")
	api.HandleFunc("/post/{id:[0-9]+}/details", handlers.UpdatePost).Methods("POST")

	api.HandleFunc("/service/clear", handlers.ClearDatabase).Methods("POST")
	api.HandleFunc("/service/status", handlers.GetDatabaseStatus).Methods("GET")

	api.HandleFunc("/thread/{slug_or_id}/create", handlers.CreatePosts).Methods("POST")
	api.HandleFunc("/thread/{slug_or_id}/details", handlers.GetThread).Methods("GET")
	api.HandleFunc("/thread/{slug_or_id}/details", handlers.UpdateThread).Methods("POST")
	api.HandleFunc("/thread/{slug_or_id}/posts", handlers.GetThreadPosts).Methods("GET")
	api.HandleFunc("/thread/{slug_or_id}/vote", handlers.VoteForPost).Methods("POST")

	api.HandleFunc("/user/{nickname}/create", handlers.CreateUser).Methods("POST")
	api.HandleFunc("/user/{nickname}/profile", handlers.GetUser).Methods("GET")
	api.HandleFunc("/user/{nickname}/profile", handlers.UpdateUser).Methods("POST")

	db := queries.InitDB("docker:docker@localhost:5432", "docker")
	defer db.Close()

	log.Println("starting server at:", 5000)
	http.ListenAndServe(":5000", r)
}
