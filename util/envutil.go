package util

import (
	"os"
	"strings"
)

func GetAnyString(names ...string) string {
	for _, n := range names {
		if strings.TrimSpace(n) != "" {
			return n
		}
	}
	return ""
}

func GetEnvAny(names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return ""
}

func GetEnvAnyWithDefault(defaultEnv string, names ...string) string {
	for _, n := range names {
		if val := os.Getenv(n); val != "" {
			return val
		}
	}
	return defaultEnv
}
