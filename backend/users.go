package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte(os.Getenv("BACKEND_SECRET"))

type returnedUser struct {
	Id            int    `json:"id"`
	DisplayName   string `json:"displayName"`
	Email         string `json:"email"`
	PublicVisible bool   `json:"publicVisisble"`
	DateCreated   string `json:"dateCreated"`
}

func (s *apiServer) handleUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		// return s.getUser(w, r)
	}
	if r.Method == "POST" {
		return s.createUser(w, r)
	}
	err := fmt.Errorf("error -> handleUser ln85. Unsuported method type %s", r.Method)
	log.Println(err.Error())
	return err
}

func (s *apiServer) handleUserByID(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	if r.Method == "GET" {
		query := `SELECT id, display_name, date_created, public_visible FROM users WHERE id=$1`
		var user returnedUser
		user.Email = "not provided with this end-point"
		row := s.store.db.QueryRow(query, id)
		err := row.Scan(&user.Id, &user.DisplayName, &user.DateCreated, &user.PublicVisible)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, user)
		return nil
	}

	if r.Method == "DELETE" {
		auth := r.Header.Get("Authorization")
		token, err := verifyToken(auth)
		if err != nil {
			return err
		}
		claims, ok := token.Claims.(jwt.MapClaims)

		if ok && token.Valid && claims["userId"] == id {
			query := `DELETE FROM users WHERE id=$1`
			res, err := s.store.db.Exec(query, id)
			if err != nil {
				return err
			}
			writeJSON(w, http.StatusOK, res)
			return nil
		}

		return fmt.Errorf("unauthorized")
	}

	if r.Method == "PUT" {

	}

	return fmt.Errorf("unsupported method")
}

func (s *apiServer) getCurrentUser(w http.ResponseWriter, r *http.Request) error {
	authToken := r.Header.Get("authorization")
	token, err := verifyToken(authToken)
	if err != nil {
		return err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		id := claims["userId"]
		query := `SELECT id, display_name, email, date_created, public_visible FROM users WHERE id=$1`

		var user returnedUser
		row := s.store.db.QueryRow(query, id)
		err = row.Scan(&user.Id, &user.DisplayName, &user.Email, &user.DateCreated, &user.PublicVisible)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, user)
		return nil
	}

	return fmt.Errorf("invalid token")
}

func (s *apiServer) loginUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not supported")
	}

	err := r.ParseForm()
	if err != nil {
		return err
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Query is case sensititve rn dont like that, fix later
	query := `SELECT id, hashed_password FROM users 
		WHERE display_name=$1 OR email=$1
	`
	row := s.store.db.QueryRow(query, username)

	var userId int
	var hashed_password string
	err = row.Scan(&userId, &hashed_password)
	if err != nil {
		return err
	}

	err = checkPasswordHash(password, hashed_password)
	if err != nil {
		return err
	}

	token, err := createToken(userId)
	if err != nil {
		return err
	}

	response := map[string]string{"auth": token}
	writeJSON(w, http.StatusOK, response)
	return nil
}

func (s *apiServer) createUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not supported")
	}

	err := r.ParseForm()
	if err != nil {
		return err
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	hashed_password, err := hashPassword(password)
	if err != nil {
		return err
	}
	displayName := r.FormValue("displayName")
	dateCreated := time.Now()
	publicVisible := true

	id := 0
	query := ` INSERT INTO users (email, display_name, date_created, public_visible, hashed_password) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = s.store.db.QueryRow(query, email, displayName, dateCreated, publicVisible, hashed_password).Scan(&id)
	if err != nil {
		return err
	}

	token, err := createToken(id)
	if err != nil {
		return err
	}

	response := map[string]string{"auth": token}
	writeJSON(w, http.StatusOK, response)
	return nil
}

func createToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userId": userId,
			"exp":    time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func checkPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
