package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"
)

// GET /
//   q     = search query (required)
//   limit = maximum number of results (optional)
//
// Example: http://localhost:8080?q=search+terms&limit=10
func queryHandler(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")
    startTime := time.Now()
    results := idx.search(q)
    limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
    if err == nil && limit > 0 && len(results) > limit {
        results = results[:limit]
    }
    if jsonResults, ok := json.Marshal(results); ok == nil {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonResults)
    } else {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Error serializing results"))
    }
    log.Printf("Processed query '%s' in %s with %d results\n",
        q, time.Since(startTime), len(results),
    )
}
