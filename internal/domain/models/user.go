package models

type User struct {
	ID             int
	Email          string
	HashedPassword []byte
	AboutText      string
	Links          []Link
}
