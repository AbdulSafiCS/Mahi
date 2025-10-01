package store

import (
	"errors"
	"sync"
	"time"
)

// Dummy user db
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type Memory struct {
	mu       sync.Mutex
	users    map[string]User
	byEmail  map[string]string // email -> id
	// refresh token store (rotation): refresh -> {userID, exp}
	refresh map[string]refreshRow
}

type refreshRow struct {
	UserID string
	Exp    time.Time
}

func NewMemory() *Memory {
	m := &Memory{
		users:   map[string]User{},
		byEmail: map[string]string{},
		refresh: map[string]refreshRow{},
	}
	// Seed one user for login demos
	u := User{ID: "u_1", Email: "demo@demo.com", Name: "Demo User"}
	m.users[u.ID] = u
	m.byEmail[u.Email] = u.ID
	return m
}

var ErrInvalidCreds = errors.New("invalid email or password")
var ErrRefreshInvalid = errors.New("invalid or expired refresh token")

func (m *Memory) VerifyCreds(email, password string) (User, error) {
	// For learning: accept any password "password" for the demo user
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byEmail[email]
	if !ok || password != "password" {
		return User{}, ErrInvalidCreds
	}
	return m.users[id], nil
}

func (m *Memory) GetUser(id string) (User, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	return u, ok
}

func (m *Memory) SaveRefresh(token, userID string, exp time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refresh[token] = refreshRow{UserID: userID, Exp: exp}
}

func (m *Memory) RotateRefresh(old string, newToken string, userID string, exp time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.refresh[old]
	if !ok || row.UserID != userID || time.Now().After(row.Exp) {
		return ErrRefreshInvalid
	}
	delete(m.refresh, old)
	m.refresh[newToken] = refreshRow{UserID: userID, Exp: exp}
	return nil
}

func (m *Memory) LookupRefresh(token string) (string, time.Time, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.refresh[token]
	return row.UserID, row.Exp, ok
}
