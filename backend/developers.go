package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type developer struct {
	ID               string `json:"id"`
	DisplayName      string `json:"displayName"`
	DateCreated      string `json:"dateCreated"`
	EntryDateCreated string `json:"entryDateCreated"`
}

func (s *apiServer) handleDeveloperByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	if r.Method == "GET" {
		var publisher publisher
		query := "SELECT (id, display_name, date_created, entry_date_created) FROM developers WHERE id=$1"
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

func (s *apiServer) handleDeveloperModerators(w http.ResponseWriter, r *http.Request) error {
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
		userId := data["userId"].(int)
		developerId := data["developerId"].(int)

		query := `INSERT INTO developer_moderators (user_id, developer_id, title) VALUES ($1, $2, $3)`
		res, err := s.store.db.Exec(query, userId, developerId, "unassociated")
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, res)
		return nil
	}

	return fmt.Errorf("unsupported method %s", r.Method)

}

func (s *apiServer) handleDeveloper(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		var returnDev developer
		var developer developer
		entryDateCreated := time.Now()
		err = json.Unmarshal(body, &developer)
		if err != nil {
			return err
		}

		query := `INSERT INTO developers (display_name, date_created, entry_date_created) VALUES ($1, $2, $3) RETURNING id, display_name, date_created, entry_date_created`
		row := s.store.db.QueryRow(query, developer.DisplayName, developer.DateCreated, entryDateCreated)
		fmt.Println(row)
		err = row.Scan(&returnDev.ID, &returnDev.DisplayName, &returnDev.DateCreated, &returnDev.EntryDateCreated)
		if err != nil {
			return err
		}

		fmt.Println(returnDev)
		writeJSON(w, http.StatusOK, returnDev)
		return nil
	}

	if r.Method == "GET" {

	}

	return fmt.Errorf("unsupported method %s", r.Method)
}
