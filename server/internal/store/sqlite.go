package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"mahi/server/internal/auth"

	_ "modernc.org/sqlite" // registers the driver
)

// SQLiteStore implements persistent storage using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// DeleteRefresh implements httpserver.Store.
func (s *SQLiteStore) DeleteRefresh(token string) error {
    _, err := s.db.Exec(`DELETE FROM refresh_tokens WHERE token=?`, token)
    return err
}



// NewSQLite opens (or creates) the DB file, runs migrations, and returns the store.
func NewSQLite(dsn string) (*SQLiteStore, error) {
	// dsn is usually a file path, e.g., "./data/app.db"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// sensible limits for a tiny app
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// migrate creates tables if they don't exist.
// Keep it simple: two tables, users and refresh_tokens.
func (s *SQLiteStore) migrate() error {
	ddl := `
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  name TEXT,
  pw_hash TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS refresh_tokens (
  token TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  exp DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
`
	_, err := s.db.Exec(ddl)
	return err
}

// ---------- Users ----------

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidCreds   = errors.New("invalid email or password")
	ErrRefreshInvalid = errors.New("invalid or expired refresh token")
)

// CreateUser inserts a new user with a temporary empty password hash.
// We'll set the real hash via SetPassword immediately after.
func (s *SQLiteStore) CreateUser(email, name string) (User, error) {
	// Generate a simple time-based id like the memory store did
	id := "u_" + time.Now().UTC().Format("20060102150405.000000000")

	// We initially write an empty pw_hash, then SetPassword will update it.
	_, err := s.db.Exec(`
INSERT INTO users (id, email, name, pw_hash) VALUES (?, ?, ?, '')
`, id, email, name)
	if err != nil {
		// SQLite returns a constraint error for duplicate emails
		if isUniqueConstraint(err) {
			return User{}, ErrEmailExists
		}
		return User{}, err
	}
	return User{ID: id, Email: email, Name: name}, nil
}

// SetPassword hashes and saves the password for a user.
func (s *SQLiteStore) SetPassword(userID, plain string) error {
	hash, err := auth.HashPassword(plain)
	if err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE users SET pw_hash = ? WHERE id = ?`, hash, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// VerifyCreds checks email + password.
func (s *SQLiteStore) VerifyCreds(email, plain string) (User, error) {
	var (
		id, name, pwHash string
	)
	row := s.db.QueryRow(`SELECT id, name, pw_hash FROM users WHERE email = ?`, email)
	if err := row.Scan(&id, &name, &pwHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidCreds
		}
		return User{}, err
	}
	if pwHash == "" || !auth.VerifyPassword(plain, pwHash) {
		return User{}, ErrInvalidCreds
	}
	return User{ID: id, Email: email, Name: name}, nil
}

func (s *SQLiteStore) GetUser(id string) (User, bool) {
	var email, name string
	row := s.db.QueryRow(`SELECT email, name FROM users WHERE id = ?`, id)
	if err := row.Scan(&email, &name); err != nil {
		return User{}, false
	}
	return User{ID: id, Email: email, Name: name}, true
}

// ---------- Refresh tokens ----------

// SaveRefresh stores/overwrites a refresh token for a user.
func (s *SQLiteStore) SaveRefresh(token, userID string, exp time.Time) {
	_, _ = s.db.Exec(`
INSERT OR REPLACE INTO refresh_tokens (token, user_id, exp) VALUES (?, ?, ?)
`, token, userID, exp.UTC())
}

// RotateRefresh deletes old and writes new atomically (best-effort).
func (s *SQLiteStore) RotateRefresh(old string, newToken string, userID string, exp time.Time) error {
	tx, err := s.db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var owner string
	row := tx.QueryRow(`SELECT user_id FROM refresh_tokens WHERE token = ?`, old)
	if err := row.Scan(&owner); err != nil {
		return ErrRefreshInvalid
	}
	if owner != userID {
		return ErrRefreshInvalid
	}
	if _, err := tx.Exec(`DELETE FROM refresh_tokens WHERE token = ?`, old); err != nil {
		return err
	}
	if _, err := tx.Exec(`
INSERT INTO refresh_tokens (token, user_id, exp) VALUES (?, ?, ?)
`, newToken, userID, exp.UTC()); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) LookupRefresh(token string) (string, time.Time, bool) {
	var userID string
	var exp time.Time
	row := s.db.QueryRow(`SELECT user_id, exp FROM refresh_tokens WHERE token = ?`, token)
	if err := row.Scan(&userID, &exp); err != nil {
		return "", time.Time{}, false
	}
	return userID, exp.UTC(), true
}

// ---------- helpers ----------

func isUniqueConstraint(err error) bool {
	// modernc.org/sqlite uses extended error codes; simplest is substring check
	// You can refine this later if needed.
	return err != nil && (contains(err.Error(), "UNIQUE") || contains(err.Error(), "unique"))
}
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool { return (stringIndex(s, sub) >= 0) })()
}
func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
