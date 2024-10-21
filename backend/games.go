package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type game struct {
	ID          int    `json:"id"`
	PublisherID int    `json:"publisherId"`
	DeveloperID int    `json:"developerId"`
	ReleaseDate string `json:"releaseDate"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var validSortOptionsGames map[string]string = map[string]string{
	"":             "id",
	"id":           "id",
	"release_date": "release_date",
	"name":         "name",
	"publisher_id": "publisher_id",
	"developer_id": "developer_id",
}

func (s *apiServer) handleGames(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		var game game
		err = json.Unmarshal(body, &game)
		if err != nil {
			return err
		}

		query := `INSERT INTO games (name, description, release_date, publisher_id, developer_id) VALUES ($1, $2, $3, $4, $5) returning id`
		row := s.store.db.QueryRow(query, game.Name, game.Description, game.ReleaseDate, game.PublisherID, game.DeveloperID)
		err = row.Scan(&game.ID)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, game)
		return nil
	}

	// EVENTUALLY MAKE THE LIST BE ABLE TO BE ASC OR DESC BASED ON QUERY
	if r.Method == "GET" {
		limitString := r.URL.Query().Get("limit")
		offsetString := r.URL.Query().Get("offset")
		sort := r.URL.Query().Get("sort")

		sort, ok := validSortOptionsGames[sort]
		if !ok {
			fmt.Println("Invalid sort option for game get query submited")
			sort = "id"
		}

		offset, err := strconv.Atoi(offsetString)
		if err != nil {
			offset = 0
		}

		limit, err := strconv.Atoi(limitString)
		if err != nil {
			query := fmt.Sprintf(`SELECT id, name, description, release_date, publisher_id, developer_id FROM games ORDER BY %s OFFSET $1`, sort)
			rows, err := s.store.db.Query(query, offset)
			if err != nil {
				return err
			}
			defer rows.Close()

			var games []game
			for rows.Next() {
				var game game
				err = rows.Scan(&game.ID, &game.Name, &game.Description, &game.ReleaseDate, &game.PublisherID, &game.DeveloperID)
				if err != nil {
					fmt.Println("there was an issue when scanning a row, gmaes, ln86")
					continue
				}
				games = append(games, game)
			}

			writeJSON(w, http.StatusOK, games)
			return nil
		}

		query := fmt.Sprintf(`SELECT id, name, description, release_date, publisher_id, developer_id FROM games ORDER BY %s LIMIT $1 OFFSET $2`, sort)
		rows, err := s.store.db.Query(query, limit, offset)
		if err != nil {
			return err
		}
		defer rows.Close()

		var games []game
		for rows.Next() {
			var game game
			err = rows.Scan(&game.ID, &game.Name, &game.Description, &game.ReleaseDate, &game.PublisherID, &game.DeveloperID)
			if err != nil {
				fmt.Println("there was an issue when scanning a row, gmaes, ln109")
				continue
			}
			games = append(games, game)
		}

		writeJSON(w, http.StatusOK, games)
		return nil
	}

	return fmt.Errorf("method not supported %s", r.Method)
}

func (s *apiServer) handleGamesByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	if r.Method == "GET" {
		var game game
		query := `SELECT id, name, description, release_date, publisher_id, developer_id FROM games WHERE id=$1`
		row := s.store.db.QueryRow(query, id)
		err := row.Scan(&game.ID, &game.Name, &game.Description, &game.ReleaseDate, &game.PublisherID, &game.DeveloperID)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, game)
		return nil
	}

	return fmt.Errorf("unsupported method")
}
