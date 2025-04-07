package query

const (
	InsertUser = "INSERT INTO users (email, username, pass_hash) VALUES($1, $2, $3) RETURNING id"

	CreatedUser = "SELECT id, email, username FROM users WHERE id = ?"

	UsersRowsByEmail = `
		SELECT 
			u.id, u.email, u.username, u.pass_hash, u.about_text,
			l.id, l.user_id, l.link_name, l.link_color, l.link_path
		FROM users u
		LEFT JOIN links l ON u.id = l.user_id
		WHERE u.email = ?`

	UsersRowsByID = `
		SELECT 
			u.id, u.email, u.username, u.pass_hash, u.about_text,
			l.id, l.user_id, l.link_name, l.link_color, l.link_path
		FROM users u
		LEFT JOIN links l ON u.id = l.user_id
		WHERE u.id = ?`

	UsersRowsByUsername = `
		SELECT 
			u.id, u.email, u.username, u.pass_hash, u.about_text,
			l.link_name, l.link_color, l.link_path
		FROM users u
		LEFT JOIN links l ON u.id = l.user_id
		WHERE u.username = ?`

	UpdateAboutMe = "UPDATE users SET about_text = ? WHERE id = ?"

	InsertLink = "INSERT INTO links (user_id, link_name, link_color, link_path) VALUES (?, ?, ?, ?)"

	ExistsLink = "SELECT EXISTS(SELECT 1 FROM links WHERE id = ? AND user_id = ?)"

	UpdateLink = "UPDATE links SET link_name = ?, link_color = ?, link_path = ? WHERE user_id = ? AND id = ?"

	DeleteLink = "DELETE FROM links WHERE id = ? AND user_id = ?"
)
