package viewModel

type LinkView struct {
	LinkName  string `json:"link_name"`
	LinkColor string `json:"link_color"`
	LinkPath  string `json:"link_path"`
}

type UserView struct {
	Username  string     `json:"username"`
	AboutText string     `json:"about"`
	Links     []LinkView `json:"links"`
}
