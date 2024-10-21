package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type postgresStore struct {
	db *sql.DB
}

func newPostgresStore() (*postgresStore, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=postgres port=5432 sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &postgresStore{
		db: db,
	}, nil
}

func (s *postgresStore) intializeDatabase() error {
	s.createUserTable()
	s.createPublisherTable()
	s.createDeveloperTable()
	s.createGameTable()
	s.createReviewTable()
	s.createTagTable()
	s.createGameTagTable()
	s.createDeveloperModeratorTable()
	s.createPublisherModeratorTable()
	s.createFriendsTable()
	return nil
}

func (s *postgresStore) createFriendsTable() error {
	query := `CREATE TABLE IF NOT EXISTS friends (
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		friend_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		status VARCHAR(10),
		PRIMARY KEY (user_id, friend_id)
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createDeveloperModeratorTable() error {
	query := `CREATE TABLE IF NOT EXISTS publisher_moderators (
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		developer_id INTEGER REFERENCES developers(id) ON DELETE CASCADE,
		title VARCHAR(30),
		PRIMARY KEY (user_id, developer_id)
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createPublisherModeratorTable() error {
	query := `CREATE TABLE IF NOT EXISTS publisher_moderators (
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		publisher_id INTEGER REFERENCES publishers(id) ON DELETE CASCADE,
		title VARCHAR(30),
		PRIMARY KEY (user_id, pubisher_id)
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createDeveloperTable() error {
	query := `CREATE TABLE IF NOT EXISTS developers (
		id SERIAL PRIMARY KEY,
		display_name VARCHAR(30) UNIQUE,
		date_created TIMESTAMP,
		entry_date_created TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createPublisherTable() error {
	query := `CREATE TABLE IF NOT EXISTS publishers (
		id SERIAL PRIMARY KEY,
		display_name VARCHAR(30) UNIQUE,
		date_created TIMESTAMP,
		entry_date_created TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createUserTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		display_name VARCHAR(12) UNIQUE,
		email VARCHAR(50) UNIQUE,
		public_visible BOOLEAN,
		date_created TIMESTAMP,
		hashed_password VARCHAR(72)
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createGameTable() error {
	query := `CREATE TABLE IF NOT EXISTS games (
		id SERIAL PRIMARY KEY,
		publisher_id INTEGER REFERENCES publishers(id),
		developer_id INTEGER REFERENCES developers(id),
		release_date TIMESTAMP,
		name VARCHAR(30),
		description TEXT
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createReviewTable() error {
	query := `CREATE TABLE IF NOT EXISTS reviews (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		game_id INTEGER REFERENCES games(id),
		rating SMALLINT,
		body TEXT,
		public_visible BOOLEAN,
		date_created TIMESTAMP
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createTagTable() error {
	query := `CREATE TABLE IF NOT EXISTS tags (
		id SERIAL PRIMARY KEY,
		tag_name VARCHAR(50)
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *postgresStore) createGameTagTable() error {
	query := `CREATE TABLE IF NOT EXISTS game_tags (
		game_id INTEGER REFERENCES games(id) ON DELETE CASCADE,
		tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
		PRIMARY KEY (game_id, tag_id)
	)`

	_, err := s.db.Exec(query)
	return err
}
