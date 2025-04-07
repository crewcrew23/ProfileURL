package requestModel

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
