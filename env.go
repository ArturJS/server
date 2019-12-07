package main

import (
	"fmt"
	"os"
)

func getenv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func mustenv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}

	return "", fmt.Errorf("ENV %s is missing", key)
}
