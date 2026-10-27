package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"github.com/Rohit-Rajvanshi/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" 
	"time"
	"github.com/google/uuid"
)


type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string 
}

type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
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




func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter , req *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w , req)
	})
}

func (cfg *apiConfig) handlerChirp(w http.ResponseWriter , req *http.Request){
	type requestBody struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	var rqBody requestBody

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rqBody)
	if err != nil {
		respondwithError(w , 500 , "Something went wrong")		
	}

	if len(rqBody.Body) > 140 {
		respondwithError(w , 400 , "Chirp is too long")
		return 
	}

	var chirp Chirp

	dbChirp , err := cfg.db.CreateChirps(req.Context() , database.CreateChirpsParams{
		Body: removeProfanities(rqBody.Body),
		UserID: rqBody.UserID,
	} )

	if err != nil {
		respondwithError(w , 500 , "Seomthing went wrong")
	}
	
	chirp.ID , chirp.CreatedAt , chirp.UpdatedAt , chirp.Body , chirp.UserID = dbChirp.ID , dbChirp.CreatedAt , dbChirp.UpdatedAt , dbChirp.Body , dbChirp.UserID

	respondwithJSON(w , 201 , chirp)

}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter , req *http.Request){
	val := cfg.fileserverHits.Load()
	strHtml := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>" , val)
	w.Header().Set("Content-Type" , "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strHtml))
}


func respondwithJSON(w http.ResponseWriter , code int , payload interface{})  error{
	response , err := json.Marshal(payload)
	if err != nil {
		return err 
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin" , "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil 
}

func removeProfanities(body string) string {

	profanities := []string{"kerfuffle" , "sharbert" , "fornax"}

	splitBody := strings.Split(body , " ")

	for i , elem := range splitBody{
		for _ , innerElem := range profanities{
			if strings.ToLower(elem) == innerElem{
				splitBody[i] = "****"
			}
		}
	}
	return strings.Join(splitBody , " ")
}

func respondwithError(w http.ResponseWriter , code int , msg string) error {
	return respondwithJSON(w , code , map[string]string{"error" : msg})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter , req *http.Request) {
	dbChirps , err := cfg.db.GetChirps(req.Context())
	if err != nil {
		respondwithError(w , 500 , "Something went wrong")
		return 
	}

	var chirps []Chirp

	for _ , elem := range dbChirps {
		chirps = append(chirps , Chirp{
			ID: elem.ID ,
			CreatedAt: elem.CreatedAt,
			UpdatedAt: elem.UpdatedAt,
			Body: elem.Body,
			UserID: elem.UserID,
		})
	}

	respondwithJSON(w , 200 , chirps)

}


func (cfg *apiConfig) handlerGetChirpByID (w http.ResponseWriter , req *http.Request){
	pathParam := req.PathValue("chirpID")
	parsedUUId , err := uuid.Parse(pathParam)
	if err != nil {
		respondwithError(w , 500 , "Error Parsing the ID")
		return 
	}

	dbChirp , err := cfg.db.GetChirpByID(req.Context() , parsedUUId)
	if err != nil {
		respondwithError(w , 404 , "Not Found")
		return 
	}

	var chirp Chirp 
	
	chirp.ID , chirp.CreatedAt , chirp.UpdatedAt , chirp.Body , chirp.UserID = dbChirp.ID , dbChirp.CreatedAt , dbChirp.UpdatedAt , dbChirp.Body , dbChirp.UserID


	respondwithJSON(w , 200 , chirp)
}


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


	serveMux.HandleFunc("POST /api/users" , func(w http.ResponseWriter , req *http.Request){
		type requestBody struct{
			Email string `json:"email"`
		}

		var rqBody requestBody

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&rqBody)
		if err != nil {
			respondwithError(w , 500 , "Something went wrong")
			return
		}

		var user User 
		dbUser , err  := apiCfg.db.CreateUser(req.Context() , rqBody.Email)
		if err != nil {
			respondwithError(w , 400 , "Bad Input")
			return 
		}

		user.ID = dbUser.ID 
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.Email = dbUser.Email

		respondwithJSON(w , 201 , user)
	})

	serveMux.Handle("/app/" , apiCfg.middlewareMetricsInc(http.StripPrefix("/app" ,http.FileServer(http.Dir(".")))))

	serveMux.Handle("/app/assets/logo.png" , http.FileServer(http.Dir(".")))

	serveMux.HandleFunc("GET /admin/metrics/" , apiCfg.handlerMetrics)

	serveMux.HandleFunc("POST /admin/reset" , apiCfg.handlerReset)

	serveMux.HandleFunc("POST /api/chirps" , apiCfg.handlerChirp)

	serveMux.HandleFunc("GET /api/chirps" , apiCfg.handlerGetChirps)

	serveMux.HandleFunc("GET /api/chirps/{chirpID}" , apiCfg.handlerGetChirpByID)

	serveMux.HandleFunc("GET /api/healthz" , func(w http.ResponseWriter , req *http.Request){
		w.Header().Set("Content-Type" , "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w , "OK")
	})

	log.Fatal(srv.ListenAndServe())
}




