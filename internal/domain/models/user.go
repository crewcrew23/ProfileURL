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

type ReqLink struct {
	LinkName  string `json:"link_name"`
	LinkColor string `json:"link_color"`
	LinkPath  string `json:"link_path"`
}

type ReqUpdateLink struct {
	LinkID    int    `json:"link_id"`
	LinkName  string `json:"link_name"`
	LinkColor string `json:"link_color"`
	LinkPath  string `json:"link_path"`
}
