package v1

import (
	"encoding/json"
	"net/http"
	"time"
)

func Healthcheck(tm time.Duration) http.HandlerFunc{
	jsonBadAns, _ := json.Marshal(map[string]interface{}{
		"status": "fall",
		"reason": "timeout",
	})
	return http.TimeoutHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ans := map[string]interface{}{
				"status": "ok",
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			jsonResponse, _ := json.Marshal(ans)
			w.Write(jsonResponse)
		},
	), tm, string(jsonBadAns)).ServeHTTP
}