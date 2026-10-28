package main 


import (
	"sync/atomic"
	"github.com/Rohit-Rajvanshi/chirpy/internal/database"
	"net/http"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string 
	secretToken string 
	polkaKey string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter , req *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w , req)
	})
}