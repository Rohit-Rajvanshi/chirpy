package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"github.com/Rohit-Rajvanshi/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" 
)



func main(){

	godotenv.Load()

	dbURL := os.Getenv("DB_URL")

	db , err := sql.Open("postgres" , dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	serveMux := http.NewServeMux()
	srv := &http.Server{
		Handler : serveMux ,
		Addr : ":8080" ,
	}

	var apiCfg apiConfig

	apiCfg.db = dbQueries
	apiCfg.platform = os.Getenv("PLATFORM")
	apiCfg.secretToken = os.Getenv("SECRET")
	apiCfg.polkaKey = os.Getenv("POLKA_KEY")

	serveMux.HandleFunc("POST /api/users" , apiCfg.handlerCreateUser)

	serveMux.Handle("/app/" , apiCfg.middlewareMetricsInc(http.StripPrefix("/app" ,http.FileServer(http.Dir(".")))))

	serveMux.Handle("/app/assets/logo.png" , http.FileServer(http.Dir(".")))

	serveMux.HandleFunc("GET /admin/metrics/" , apiCfg.handlerMetrics)

	serveMux.HandleFunc("POST /admin/reset" , apiCfg.handlerReset)

	serveMux.HandleFunc("POST /api/chirps" , apiCfg.handlerChirp)

	serveMux.HandleFunc("GET /api/chirps" , apiCfg.handlerGetChirps)

	serveMux.HandleFunc("GET /api/chirps/{chirpID}" , apiCfg.handlerGetChirpByID)

	serveMux.HandleFunc("POST /api/login/" , apiCfg.handlerLogin)

	serveMux.HandleFunc("POST /api/refresh" , apiCfg.handlerRefresh)

	serveMux.HandleFunc("POST /api/revoke" , apiCfg.handlerRevoke)

	serveMux.HandleFunc("PUT /api/users" , apiCfg.handlerUpdateUser)

	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}" , apiCfg.handlerDelete)
	
	serveMux.HandleFunc("POST /api/polka/webhooks" , apiCfg.handlerWebhook)

	serveMux.HandleFunc("GET /api/healthz" , handlerHealthz)

	log.Fatal(srv.ListenAndServe())
}




