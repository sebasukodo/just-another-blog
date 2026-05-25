package config

import "os"

type Env struct {
	DBName      string
	DBHost      string
	DBUser      string
	DBPassword  string
	TokenSecret string
	TokenIssuer string
	LogFile     string
}

func Load() *Env {
	return &Env{
		DBName:      os.Getenv("DB_NAME"),
		DBHost:      os.Getenv("DB_HOST"),
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		TokenSecret: os.Getenv("TOKEN_SECRET"),
		TokenIssuer: os.Getenv("TOKEN_ISSUER"),
		LogFile:     os.Getenv("LOG_FILE"),
	}
}
