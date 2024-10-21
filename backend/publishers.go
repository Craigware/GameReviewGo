package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type publisher struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName"`
	DateCreated      string `json:"dateCreated"`
	EntryDateCreated string `json:"entryDateCreated"`
}

func (s *apiServer) handlePublisherByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	if r.Method == "GET" {
		var publisher publisher
		query := "SELECT (id, display_name, date_created, entry_date_created) FROM publishers WHERE id=$1"
		row := s.store.db.QueryRow(query, id)
		err := row.Scan(&publisher.ID, &publisher.DisplayName, &publisher.DateCreated, &publisher.EntryDateCreated)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, publisher)
		return nil
	}

	return fmt.Errorf("unsupported method %s", r.Method)
}

func (s *apiServer) handlePublisherModerators(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		defer r.Body.Close()
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			return err
		}
		userId := data["userId"].(string)
		publisherId := data["publisherId"].(string)

		query := `INSERT INTO publisher_moderators (user_id, publisher_id, title) VALUES ($1, $2, $3)`
		res, err := s.store.db.Exec(query, userId, publisherId, "unassociated")
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, res)
		return nil
	}

	if r.Method == "GET" {

	}

	return fmt.Errorf("unsupported method %s", r.Method)
}

func (s *apiServer) handlePublisher(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		var returnPub publisher
		var publisher publisher
		entryDateCreated := time.Now()
		err = json.Unmarshal(body, &publisher)
		if err != nil {
			return err
		}

		query := `INSERT INTO publishers (display_name, date_created, entry_date_created) VALUES ($1, $2, $3) RETURNING id, display_name, date_created, entry_date_created`
		err = s.store.db.QueryRow(query, &publisher.DisplayName, &publisher.DateCreated, entryDateCreated).Scan(&returnPub.ID, &returnPub.DisplayName, &returnPub.DateCreated, &returnPub.EntryDateCreated)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, returnPub)
		return nil
	}

	return fmt.Errorf("unsupported method %s", r.Method)
}
