package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiServer struct {
	addr  string
	store *postgresStore
}

func main() {
	fmt.Println("Server starting...")
	server := newApiServer(":8000")
	server.run()
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
	}
}

func newApiServer(listenAddr string) *apiServer {
	store, err := newPostgresStore()
	if err != nil {
		log.Fatal(err)
	}
	store.intializeDatabase()

	return &apiServer{
		addr:  listenAddr,
		store: store,
	}
}

func (s *apiServer) run() {
	router := mux.NewRouter()

	// USERS
	router.HandleFunc("/users", makeHTTPHandleFunc(s.handleUser)).Methods("POST", "GET")
	router.HandleFunc("/users/login", makeHTTPHandleFunc(s.loginUser)).Methods("POST")
	router.HandleFunc("/users/current", makeHTTPHandleFunc(s.getCurrentUser)).Methods("GET")
	router.HandleFunc("/users/{id}", makeHTTPHandleFunc(s.handleUserByID)).Methods("GET", "DELETE", "PATCH", "PUT")

	// FRIENDS
	router.HandleFunc("/users/{id}/friend_request", makeHTTPHandleFunc(s.handleFriendRequest))
	router.HandleFunc("/friends", makeHTTPHandleFunc(s.handleFriends)).Methods("GET")

	// PUBLISHERS
	router.HandleFunc("/publishers", makeHTTPHandleFunc(s.handlePublisher)).Methods("POST", "GET")
	router.HandleFunc("/publishers/{id}", makeHTTPHandleFunc(s.handlePublisherByID)).Methods("GET", "PUT", "PATCH", "DELETE")
	router.HandleFunc("/publishers/{id}/moderators", makeHTTPHandleFunc(s.handlePublisherModerators)).Methods("POST", "DELETE", "GET")

	// DEVELOPERS
	router.HandleFunc("/developers", makeHTTPHandleFunc(s.handleDeveloper)).Methods("POST", "GET")
	router.HandleFunc("/developers/{id}", makeHTTPHandleFunc(s.handleDeveloperByID)).Methods("GET", "PUT", "PATCH", "DELETE")
	router.HandleFunc("/developers/{id}/moderators", makeHTTPHandleFunc(s.handleDeveloperModerators)).Methods("POST", "DELETE", "GET")

	// GAMES
	router.HandleFunc("/games", makeHTTPHandleFunc(s.handleGames))
	router.HandleFunc("/games/{id}", makeHTTPHandleFunc(s.handleGamesByID))

	// REVIEWS
	router.HandleFunc("/games/{game_id}/reviews", makeHTTPHandleFunc(s.handleReviewsByGameID))
	// router.HandleFunc("/users/{id}/reviews")

	log.Println("Backend -> Unnamed Game Review Site, running. Port:", s.addr)
	http.ListenAndServe(s.addr, router)
}
