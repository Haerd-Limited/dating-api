package request

import (
	"net/http"
	"strconv"
)

func ParseQueryInt(r *http.Request, key string, defaultValue int) int {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(valStr)
	if err != nil || val < 0 {
		return defaultValue
	}

	return val
}
