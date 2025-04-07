package models

type User struct {
	ID             int
	Email          string
	Username       string
	HashedPassword []byte
	AboutText      string
	Links          []Link
}
