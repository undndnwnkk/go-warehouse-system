package main

import (
	"os"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBURL      string
	DBName     string
	JWTSecret  string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
