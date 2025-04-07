CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    pass_hash BLOB NOT NULL,
    about_text TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS links(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    link_name TEXT NOT NULL DEFAULT '',
    link_color TEXT NOT NULL DEFAULT '',
    link_path TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_links_user_id ON links (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (email);