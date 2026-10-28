package main

import (
	"net/http"
	"fmt"
	"io"
)

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter , req *http.Request){
	val := cfg.fileserverHits.Load()
	strHtml := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>" , val)
	w.Header().Set("Content-Type" , "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strHtml))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter , req *http.Request){

	if cfg.platform != "dev" {
		respondwithError(w , 403 , "Forbidden")
		return
	}

	err := cfg.db.ResetUsers(req.Context())
	if err != nil {
		respondwithError(w , 400 , "Something went wrong")
		return
	}
}

func handlerHealthz (w http.ResponseWriter , req *http.Request){
	w.Header().Set("Content-Type" , "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w , "OK")
}