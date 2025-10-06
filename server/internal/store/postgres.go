package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"mahi/server/internal/auth"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx as database/sql driver
)

type Postgres struct {
    db *sql.DB
}

func NewPostgres(dsn string) (*Postgres, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, err
    }
    // Pool tuning: adjust as needed
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(30 * time.Minute)

    p := &Postgres{db: db}
    if err := p.migrate(); err != nil {
        return nil, err
    }
    return p, nil
}

func (p *Postgres) migrate() error {
    _, err := p.db.Exec(`
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT,
  pw_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS refresh_tokens (
  token TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  exp_unix BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id);
`)
    return err
}

func (p *Postgres) CreateUser(email, name string) (User, error) {
    id := "u_" + time.Now().UTC().Format("20060102150405.000000000")
    // insert with placeholder pw so row exists; SetPassword updates it
    _, err := p.db.Exec(`
        INSERT INTO users (id,email,name,pw_hash) VALUES ($1,$2,$3,$4)
    `, id, email, name, "placeholder")
    if err != nil {
        if isPGUnique(err) {
            return User{}, ErrEmailExists
        }
        return User{}, err
    }
    return User{ID: id, Email: email, Name: name}, nil
}

func (p *Postgres) SetPassword(userID, plain string) error {
    hash, err := auth.HashPassword(plain)
    if err != nil {
        return err
    }
    res, err := p.db.Exec(`UPDATE users SET pw_hash=$1, updated_at=now() WHERE id=$2`, hash, userID)
    if err != nil {
        return err
    }
    n, _ := res.RowsAffected()
    if n == 0 {
        return ErrUserNotFound
    }
    return nil
}

func (p *Postgres) VerifyCreds(email, password string) (User, error) {
    var id, name, pwHash string
    err := p.db.QueryRow(`SELECT id, COALESCE(name,''), pw_hash FROM users WHERE email=$1`, email).
        Scan(&id, &name, &pwHash)
    if errors.Is(err, sql.ErrNoRows) {
        return User{}, ErrInvalidCreds
    }
    if err != nil {
        return User{}, err
    }
    if !auth.VerifyPassword(password, pwHash) {
        return User{}, ErrInvalidCreds
    }
    return User{ID: id, Email: email, Name: name}, nil
}

func (p *Postgres) GetUser(id string) (User, bool) {
    var email, name string
    err := p.db.QueryRow(`SELECT email, COALESCE(name,'') FROM users WHERE id=$1`, id).
        Scan(&email, &name)
    if errors.Is(err, sql.ErrNoRows) || err != nil {
        return User{}, false
    }
    return User{ID: id, Email: email, Name: name}, true
}

func (p *Postgres) SaveRefresh(token, userID string, exp time.Time) {
    _, _ = p.db.Exec(`INSERT INTO refresh_tokens (token,user_id,exp_unix) VALUES ($1,$2,$3)
                      ON CONFLICT (token) DO UPDATE SET user_id=EXCLUDED.user_id, exp_unix=EXCLUDED.exp_unix`,
        token, userID, exp.Unix())
}

func (p *Postgres) RotateRefresh(old, newToken, userID string, exp time.Time) error {
    ctx := context.Background()
    tx, err := p.db.BeginTx(ctx, nil)
    if err != nil { return err }
    defer func() { _ = tx.Rollback() }()

    var uid string
    var expUnix int64
    err = tx.QueryRowContext(ctx, `SELECT user_id, exp_unix FROM refresh_tokens WHERE token=$1`, old).
        Scan(&uid, &expUnix)
    if errors.Is(err, sql.ErrNoRows) { return ErrRefreshInvalid }
    if err != nil { return err }
    if uid != userID || time.Now().Unix() > expUnix { return ErrRefreshInvalid }

    if _, err := tx.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE token=$1`, old); err != nil {
        return err
    }
    if _, err := tx.ExecContext(ctx, `INSERT INTO refresh_tokens (token,user_id,exp_unix) VALUES ($1,$2,$3)`,
        newToken, userID, exp.Unix()); err != nil {
        return err
    }
    return tx.Commit()
}

func (p *Postgres) LookupRefresh(token string) (string, time.Time, bool) {
    var uid string
    var expUnix int64
    err := p.db.QueryRow(`SELECT user_id, exp_unix FROM refresh_tokens WHERE token=$1`, token).
        Scan(&uid, &expUnix)
    if errors.Is(err, sql.ErrNoRows) || err != nil {
        return "", time.Time{}, false
    }
    return uid, time.Unix(expUnix, 0), true
}

func (p *Postgres) DeleteRefresh(token string) error {
    _, err := p.db.Exec(`DELETE FROM refresh_tokens WHERE token=$1`, token)
    return err
}

// crude unique violation detector (pgx via database/sql encodes codes on err string)
func isPGUnique(err error) bool {
    if err == nil { return false }
    s := err.Error()
    // Postgres code 23505 = unique_violation
    return (len(s) > 0 && (contains(s, "23505") || contains(s, "unique constraint")))
}
