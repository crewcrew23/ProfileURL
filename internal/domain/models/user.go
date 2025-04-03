package models

type User struct {
	ID             int
	Email          string
	HashedPassword []byte
	AboutText      string
	Links          []Link
}

type Link struct {
	ID        int
	UserID    int
	LinkName  string
	LinkColor string
	LinkPath  string
}
