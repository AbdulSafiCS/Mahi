package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	JWTSecret      string
	AccessTTLMin   int
	RefreshTTLDays int
	DBPath          string
}

// read env variables. set default if not set. 
func Load() Config {
	_ = godotenv.Load(".env")

	return Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "change_me"),
		AccessTTLMin:   getEnvInt("ACCESS_TTL_MIN", 15),
		RefreshTTLDays: getEnvInt("REFRESH_TTL_DAYS", 30),
		DBPath:         getEnv("DB_PATH", "data/app.db"),
	}
}
// helper function - checks Getenv and parses ints safely 
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
// helper function - checks Getenv and parses ints safely 
func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
