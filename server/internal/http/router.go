package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"mahi/server/internal/auth"
	"mahi/server/internal/config"
	"mahi/server/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type Store interface {
    CreateUser(email, name string) (store.User, error)
    SetPassword(userID, plain string) error
    VerifyCreds(email, password string) (store.User, error)
    GetUser(id string) (store.User, bool)
    SaveRefresh(token, userID string, exp time.Time)
    RotateRefresh(old, newToken, userID string, exp time.Time) error
    LookupRefresh(token string) (string, time.Time, bool)
	DeleteRefresh(token string) error
}

type Server struct {
    cfg config.Config
    jwt *auth.JWTMaker
    st  Store // <-- use the interface instead of *store.Memory
}

func NewRouter(cfg config.Config) http.Handler {
	  var st Store
    var err error

    switch cfg.DBDriver {
    case "postgres":
        st, err = store.NewPostgres(cfg.DBDSN)
    case "sqlite":
        fallthrough
    default:
        st, err = store.NewSQLite(cfg.DBPath)
    }
    if err != nil {
        panic(err)
    }

    s := &Server{
        cfg: cfg,
        jwt: auth.NewJWTMaker(cfg.JWTSecret),
        st:  st,
    }

	r := chi.NewRouter()

	// Dev CORS — loosen for now
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET","POST","PUT","DELETE","OPTIONS"},
		AllowedHeaders:   []string{"Authorization","Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health
	r.Get("/healthz", s.health)

	// Auth
	r.Route("/v1", func(r chi.Router) {
		 r.Post("/auth/register", s.register)
		r.Post("/auth/login", s.login)
		r.Post("/auth/refresh", s.refresh)
		r.Post("/auth/logout", s.logout) 

		r.Group(func(pr chi.Router) {
			pr.Use(s.authn) // JWT middleware
			pr.Get("/users/me", s.me)
		})
	})

	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok"})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type tokenResp struct {
	AccessToken     string      `json:"access_token"`
	AccessExpiresIn int         `json:"access_expires_in"`
	RefreshToken    string      `json:"refresh_token"`
	RefreshExpiresIn int        `json:"refresh_expires_in"`
	User            store.User  `json:"user"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}
	u, err := s.st.VerifyCreds(req.Email, req.Password)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid_credentials", nil)
		return
	}

	// Access token
	access, _, err := s.jwt.NewAccess(u.ID, s.cfg.AccessTTLMin)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token_error", nil)
		return
	}

	// Refresh token (opaque string for demo)
	rt := newRefreshToken()
	rtExp := time.Now().Add(time.Duration(s.cfg.RefreshTTLDays) * 24 * time.Hour)
	s.st.SaveRefresh(rt, u.ID, rtExp)

	writeJSON(w, http.StatusOK, tokenResp{
		AccessToken: access,
		AccessExpiresIn: s.cfg.AccessTTLMin*60,
		RefreshToken: rt,
		RefreshExpiresIn: int(time.Until(rtExp).Seconds()),
		User: u,
	})
}
// POST /v1/auth/register
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid_json", nil)
        return
    }
    if req.Email == "" || req.Password == "" {
        writeErr(w, http.StatusBadRequest, "missing_fields", nil)
        return
    }

    // 1) Create user (fails if email exists)
    u, err := s.st.CreateUser(req.Email, req.Name)
    if err != nil {
        // expect something like store.ErrEmailExists; fall back to 409
        writeErr(w, http.StatusConflict, "email_exists", nil)
        return
    }

    // 2) Hash & set password
    if err := s.st.SetPassword(u.ID, req.Password); err != nil {
        writeErr(w, http.StatusInternalServerError, "password_set_failed", nil)
        return
    }

    // 3) Issue access token (JWT)
    access, _, err := s.jwt.NewAccess(u.ID, s.cfg.AccessTTLMin)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "token_error", nil)
        return
    }

    // 4) Issue refresh token (opaque) and persist it
    rt := newRefreshToken()
    rtExp := time.Now().Add(time.Duration(s.cfg.RefreshTTLDays) * 24 * time.Hour)
    s.st.SaveRefresh(rt, u.ID, rtExp)

    // 5) Respond (201 Created)
    writeJSON(w, http.StatusCreated, tokenResp{
        AccessToken:       access,
        AccessExpiresIn:   s.cfg.AccessTTLMin * 60,
        RefreshToken:      rt,
        RefreshExpiresIn:  int(time.Until(rtExp).Seconds()),
        User:              u,
    })
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}
type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", nil)
		return
	}
	userID, exp, ok := s.st.LookupRefresh(req.RefreshToken)
	if !ok || time.Now().After(exp) {
		writeErr(w, http.StatusUnauthorized, "refresh_invalid", nil)
		return
	}
	// new access
	access, _, err := s.jwt.NewAccess(userID, s.cfg.AccessTTLMin)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token_error", nil)
		return
	}
	// rotate refresh
	newRT := newRefreshToken()
	newExp := time.Now().Add(time.Duration(s.cfg.RefreshTTLDays) * 24 * time.Hour)
	if err := s.st.RotateRefresh(req.RefreshToken, newRT, userID, newExp); err != nil {
		writeErr(w, http.StatusUnauthorized, "refresh_invalid", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": access,
		"access_expires_in": s.cfg.AccessTTLMin*60,
		"refresh_token": newRT,
		"refresh_expires_in": int(time.Until(newExp).Seconds()),
	})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ctxKeyUserID{}).(string)
	u, ok := s.st.GetUser(userID)
	if !ok {
		writeErr(w, http.StatusNotFound, "user_not_found", nil)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

type logoutReq struct {
    RefreshToken string `json:"refresh_token"`
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
    var req logoutReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid_json", nil); return
    }
    if req.RefreshToken == "" { writeErr(w, http.StatusBadRequest, "missing_token", nil); return }
    _ = s.st.DeleteRefresh(req.RefreshToken)
    writeJSON(w, http.StatusOK, map[string]string{"status":"ok"})
}
