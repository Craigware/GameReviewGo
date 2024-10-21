package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type friend struct {
	UserID   int    `json:"userId"`
	FriendID int    `json:"friendId"`
	Status   string `json:"status"`
}

func (s *apiServer) handleFriendRequest(w http.ResponseWriter, r *http.Request) error {
	var id, currentId int
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return err
	}

	authToken := r.Header.Get("authorization")
	token, err := verifyToken(authToken)
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("error with authentication token")
	}
	floatId, ok := claims["userId"].(float64)
	if !ok {
		return fmt.Errorf("something wrong with authorization claims")
	}
	currentId = int(floatId)

	if r.Method == "POST" {
		var friend friend
		friend.FriendID = id
		friend.UserID = currentId
		query := `SELECT status FROM friends WHERE (friend_id=$1 AND user_id=$2) OR (friend_id=$2 AND user_id=$1)`
		row := s.store.db.QueryRow(query, currentId, id)
		err = row.Scan()
		if err != sql.ErrNoRows {
			return fmt.Errorf("a friend status for these users already exists in the database")
		}

		query = "INSERT INTO friends (user_id, friend_id, status) VALUES ($1, $2, $3) RETURNING status"
		row = s.store.db.QueryRow(query, currentId, id, "PENDING")
		err = row.Scan(&friend.Status)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, friend)
		return nil
	}

	if r.Method == "PATCH" {
		var friend friend
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &friend)
		if err != nil {
			return err
		}

		query := "UPDATE friends SET status=$1 WHERE (friend_id=$2 AND user_id=$3) RETURNING friend_id, user_id"
		row := s.store.db.QueryRow(query, &friend.Status, currentId, id)
		err = row.Scan(&friend.FriendID, &friend.UserID)
		if err != nil {
			return fmt.Errorf("either a friend status does not exist in the database or you tried to accept the friend request from the wrong side")
		}

		writeJSON(w, http.StatusOK, friend)
		return nil
	}

	if r.Method == "DELETE" {
		var friend friend
		friend.Status = "DECLINED"

		query := "DELETE FROM friends WHERE (friend_id=$1 AND user_id=$2) OR (friend_id=$2 AND user_id=$1) RETURNING friend_id, user_id"
		row := s.store.db.QueryRow(query, currentId, id)
		err = row.Scan(&friend.FriendID, &friend.UserID)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, friend)
		return nil
	}

	return fmt.Errorf("unsupported method %s", r.Method)
}

func (s *apiServer) handleFriends(w http.ResponseWriter, r *http.Request) error {
	authToken := r.Header.Get("authorization")
	token, err := verifyToken(authToken)
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("error with authentication token")
	}

	floatId, ok := claims["userId"].(float64)
	if !ok {
		return fmt.Errorf("something wrong with authorization claims")
	}

	id := int(floatId)

	if r.Method == "GET" {
		var friends []friend
		query := `SELECT user_id, friend_id, status FROM friends WHERE friend_id=$1 OR user_id=$1 AND status='ACCEPTED'`
		rows, err := s.store.db.Query(query, id)
		if err != nil {
			return nil
		}
		defer rows.Close()

		for rows.Next() {
			var friend friend
			err = rows.Scan(&friend.UserID, &friend.FriendID, &friend.Status)
			if err != nil {
				fmt.Println("an error occured while getting friend status' ln114")
				continue
			}

			friends = append(friends, friend)
		}

		writeJSON(w, http.StatusOK, friends)
		return nil
	}

	return fmt.Errorf("method not supported %s", r.Method)
}
