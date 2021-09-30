package main

import (
	"os"
	"strconv"
	"strings"
)

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val, ok := os.LookupEnv(key); ok == true {
		if valInt, err := strconv.Atoi(val); err == nil {
			return valInt
		}
	}
	return defaultValue
}

func getEnvSlice(key string) []string {
	val := getEnv(key, "")
	if val != "" {
		return strings.Split(val, ",")
	}
	return nil
}
