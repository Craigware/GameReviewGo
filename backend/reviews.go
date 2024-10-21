package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type review struct {
	ID            int    `json:"id"`
	UserID        int    `json:"userId"`
	GameId        int    `json:"gameId"`
	Rating        int    `json:"rating"`
	Body          string `json:"body"`
	PublicVisible bool   `json:"publicVisible"`
	DateCreated   string `json:"dateCreated"`
}

var validSortOptionsReviews map[string]string = map[string]string{
	"":             "id",
	"id":           "id",
	"user_id":      "user_id",
	"rating":       "rating",
	"date_created": "date_created",
}

// func (s *apiServer) handleReviewsByID(w http.ResponseWriter, r *http.Request) error {
// 	return fmt.Errorf("unsupported method %s", r.Method)
// }

func (s *apiServer) handleReviewsByGameID(w http.ResponseWriter, r *http.Request) error {
	gameId := mux.Vars(r)["game_id"]
	var userId int
	auth := r.Header.Get("Authorization")
	token, err := verifyToken(auth)
	if err == nil {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			userId = int(claims["userId"].(float64))
		}
	}

	// Need to create the friends status table before making non public visible visible to friends
	if r.Method == "GET" {
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")
		sort := r.URL.Query().Get("sort")

		sort, ok := validSortOptionsReviews[sort]
		if !ok {
			fmt.Println("Invalid sort option for review get query submited")
			sort = "id"
		}

		gameIdInt, err := strconv.Atoi(gameId)
		if err != nil {
			return err
		}

		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			offsetInt = 0
		}

		var reviews []review
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			if userId == 0 {
				reviews, err = s.getGameReviewsOS(gameIdInt, offsetInt, sort)
				if err != nil {
					return err
				}
			} else {
				reviews, err = s.getGameReviewsAOS(gameIdInt, offsetInt, sort, userId)
				if err != nil {
					return err
				}
			}
			writeJSON(w, http.StatusOK, reviews)
			return nil
		}

		if userId == 0 {
			reviews, err = s.getGameReviewsLOS(gameIdInt, offsetInt, limitInt, sort)
			if err != nil {
				return err
			}
		} else {
			reviews, err = s.getGameReviewsALOS(gameIdInt, offsetInt, limitInt, sort, userId)
			if err != nil {
				return err
			}
		}

		writeJSON(w, http.StatusOK, reviews)
		return nil
	}

	if r.Method == "POST" {
		if userId == 0 {
			return fmt.Errorf("user not authorized")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		var review review
		err = json.Unmarshal(body, &review)
		if err != nil {
			return err
		}

		review.UserID = userId
		review.GameId, err = strconv.Atoi(gameId)
		if err != nil {
			return err
		}

		dateCreated := time.Now()

		query := `INSERT INTO reviews (user_id, game_id, rating, body, public_visible, date_created) VALUES ($1, $2, $3, $4, $5, $6) returning id, date_created`
		row := s.store.db.QueryRow(query, &review.UserID, &review.GameId, &review.Rating, &review.Body, &review.PublicVisible, dateCreated)
		err = row.Scan(&review.ID, &review.DateCreated)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, review)
		return nil
	}

	return fmt.Errorf("unsupported method %s", r.Method)
}

func (s *apiServer) getGameReviewsOS(gameIdInt int, offset int, sort string) ([]review, error) {
	query := fmt.Sprintf(`SELECT id, user_id, rating, body, date_created FROM reviews WHERE game_id=$1 AND public_visible ORDER BY %s OFFSET $2`, sort)
	rows, err := s.store.db.Query(query, gameIdInt, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []review
	for rows.Next() {
		var review review
		review.GameId = gameIdInt
		err = rows.Scan(&review.ID, &review.UserID, &review.Rating, &review.Body, &review.DateCreated)
		if err != nil {
			fmt.Println("error while scanning reviews")
			continue
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (s *apiServer) getGameReviewsLOS(gameIdInt int, offset int, limit int, sort string) ([]review, error) {
	query := fmt.Sprintf(`SELECT id, user_id, rating, body, date_created FROM reviews WHERE game_id=$1 AND public_visible ORDER BY %s LIMIT $2 OFFSET $3`, sort)
	rows, err := s.store.db.Query(query, gameIdInt, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []review
	for rows.Next() {
		var review review
		review.GameId = gameIdInt
		err = rows.Scan(&review.ID, &review.UserID, &review.Rating, &review.Body, &review.DateCreated)
		if err != nil {
			fmt.Println("error while scanning reviews")
			continue
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (s *apiServer) getGameReviewsALOS(gameIdInt int, offset int, limit int, sort string, userId int) ([]review, error) {
	query := fmt.Sprintf(`SELECT r.id, r.user_id, r.rating, r.body, r.date_created FROM reviews r
	LEFT JOIN friends f ON ((f.user_id=$4 AND r.user_id=f.friend_id) OR (f.friend_id=$4 AND r.user_id=f.user_id))
	WHERE game_id=$1 
	AND (public_visible OR ((f.user_id=$3 OR f.friend_id=$3) AND f.status='ACCEPTED')) OR r.user_id=$4
	ORDER BY %s LIMIT $2 OFFSET $3`, sort)
	rows, err := s.store.db.Query(query, gameIdInt, limit, offset, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []review
	for rows.Next() {
		var review review
		review.GameId = gameIdInt
		err = rows.Scan(&review.ID, &review.UserID, &review.Rating, &review.Body, &review.DateCreated)
		if err != nil {
			fmt.Println("error while scanning reviews")
			continue
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func (s *apiServer) getGameReviewsAOS(gameIdInt int, offset int, sort string, userId int) ([]review, error) {
	query := fmt.Sprintf(`SELECT r.id, r.user_id, r.rating, r.body, r.date_created FROM reviews r
		LEFT JOIN friends f ON ((f.user_id=$3 AND r.user_id=f.friend_id) OR (f.friend_id=$3 AND r.user_id=f.user_id))
		WHERE game_id=$1 
		AND (r.public_visible OR ((f.user_id=$3 OR f.friend_id=$3) AND f.status='ACCEPTED')) OR r.user_id=$3
		ORDER BY %s OFFSET $2`, sort)
	rows, err := s.store.db.Query(query, gameIdInt, offset, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []review
	for rows.Next() {
		var review review
		review.GameId = gameIdInt
		err = rows.Scan(&review.ID, &review.UserID, &review.Rating, &review.Body, &review.DateCreated)
		if err != nil {
			fmt.Println("error while scanning reviews")
			continue
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}
