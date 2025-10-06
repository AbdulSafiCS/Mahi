-- users table
CREATE TABLE IF NOT EXISTS users (
  id        TEXT PRIMARY KEY,
  email     TEXT NOT NULL UNIQUE,
  name      TEXT,
  pw_hash   TEXT NOT NULL,              -- argon2id hash
  created_at DATETIME NOT NULL DEFAULT (datetime('now')),
  updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

-- refresh tokens (rotate)
CREATE TABLE IF NOT EXISTS refresh_tokens (
  token     TEXT PRIMARY KEY,
  user_id   TEXT NOT NULL,
  exp_unix  INTEGER NOT NULL,           -- expiration epoch seconds
  created_at DATETIME NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id);
