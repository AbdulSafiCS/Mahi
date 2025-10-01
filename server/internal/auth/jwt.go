package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// HMAC secret
type JWTMaker struct {
	secret []byte
}
// parse and sign JWT tokens
func NewJWTMaker(secret string) *JWTMaker {
	return &JWTMaker{secret: []byte(secret)}
}

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}
// Builds a new HS256 JWT with userID and an expiry. Returns the token string and the expiry time.
func (j *JWTMaker) NewAccess(userID string, ttlMin int) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(time.Duration(ttlMin) * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   userID,
		},
	})
	s, err := token.SignedString(j.secret)
	return s, exp, err
}
// validate or return claim if expired or error
func (j *JWTMaker) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := token.Claims.(*Claims); ok && token.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
