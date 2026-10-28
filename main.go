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
	"github.com/Rohit-Rajvanshi/chirpy/internal/auth"
	"sort"
)


type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string 
	secretToken string 
	polkaKey string
}

type User struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token 	  string    `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed bool `json:"is_chirpy_red"`
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

	token , err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondwithError(w , 401 , "Error getting the token")
		return 
	}

	authID , err := auth.ValidateJWT(token , cfg.secretToken)
	if err != nil {
		respondwithError(w , 401 , "Error validating the token")
		return 
	}


	var chirp Chirp

	dbChirp , err := cfg.db.CreateChirps(req.Context() , database.CreateChirpsParams{
		Body: removeProfanities(rqBody.Body),
		UserID: authID,
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

	authorID := req.URL.Query().Get("author_id")
	sortMethod := req.URL.Query().Get("sort")

	if sortMethod == ""{
		sortMethod = "asc"
	}




	

	Chirps := []database.Chirp{}

	if authorID != "" {
		parsedUUId , err := uuid.Parse(authorID)
		if err != nil {
			respondwithError(w , 500 , "Error Parsing the ID")
			return
		}
		dbChirps , err := cfg.db.GetChirpsForUser(req.Context() , parsedUUId)
		if err != nil {
			respondwithError(w , 500 , "Something went wrong")
			return 
		}
		Chirps = dbChirps
	}else{
		dbChirps , err := cfg.db.GetChirps(req.Context())
		if err != nil {
			respondwithError(w , 500 , "Something went wrong")
			return 
		}
		Chirps = dbChirps
	}

	var chirps []Chirp

	for _ , elem := range Chirps {
		chirps = append(chirps , Chirp{
			ID: elem.ID ,
			CreatedAt: elem.CreatedAt,
			UpdatedAt: elem.UpdatedAt,
			Body: elem.Body,
			UserID: elem.UserID,
		})
	}

	if sortMethod == "asc" {
		sort.Slice(chirps , func(i , j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}else{
		sort.Slice(chirps , func(i , j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}


	respondwithJSON(w , 200 , chirps )

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

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter , req *http.Request) {
	type requestBody struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}

	var rqBody requestBody

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rqBody)
	if err != nil {
		respondwithError(w , 500 , "Something went wrong")
		return
	}


	dbUser , err := cfg.db.GetUserByEmail(req.Context() , rqBody.Email)
	if err != nil {
		respondwithError(w , 401 , "Incorrect Email or Password")
		return
	}

	match , err := auth.CheckPasswordHash(rqBody.Password , dbUser.HashedPassword)
	if err != nil {
		respondwithError(w , 500 , "Error")
		return 
	}

	if match != true {
		respondwithError(w , 401 , "Incorrect email or password")
		return
	}

	expiryTime := 3600 * time.Second 

	var user User 
	user.ID = dbUser.ID 
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	user.Email = dbUser.Email
	user.IsChirpyRed = dbUser.IsChirpyRed
	user.Token , err = auth.MakeJWT(user.ID , cfg.secretToken , expiryTime )
	if err != nil {
		respondwithError(w , 500 , "Error Creating Token")
		return
	}
	user.RefreshToken = auth.MakeRefreshToken()
	_ , err = cfg.db.CreateRefreshToken(req.Context() , database.CreateRefreshTokenParams{
		Token : user.RefreshToken ,
		UserID: user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour ),
	})
	if err != nil {
		respondwithError(w , 500 , "Error Saving the Refresh Token in DB")
		return 
	}

	respondwithJSON(w , 200 , user)
}

func (cfg *apiConfig) handlerRefresh (w http.ResponseWriter , req *http.Request){
	
	type responseBody struct {
		Token string `json:"token"`
	}

	ref_token , err  := auth.GetBearerToken(req.Header)
	if err != nil {
		respondwithError(w , 401 , "Invalid Header")
		return 
	}

	dbRefToken , err := cfg.db.GetToken(req.Context() , ref_token)
	if err != nil {
		respondwithError(w , 401 , "Invalid Token")
		return 
	}

	dbUser , err := cfg.db.GetUserFromRefreshToken( req.Context()  , dbRefToken.UserID)
	if err != nil {
		respondwithError(w , 401 , "User not Found")
		return 
	}

	var rspBody responseBody

	token , err  := auth.MakeJWT(dbUser.ID , cfg.secretToken , 3600 * time.Second)
	if err != nil {
		respondwithError(w , 500 , "Something went wrong")
		return 
	}

	rspBody.Token = token

	respondwithJSON(w , 200 , rspBody)
}

func (cfg *apiConfig) handlerRevoke (w http.ResponseWriter , req *http.Request) {
	ref_token , err  := auth.GetBearerToken(req.Header)
	if err != nil {
		respondwithError(w , 401 , "Invalid Header")
		return 
	}

	dbRefToken , err := cfg.db.GetToken(req.Context() , ref_token)
	if err != nil {
		respondwithError(w , 401 , "Invalid Token")
		return 
	}

	RevokeErr := cfg.db.RevokeToken(req.Context() , dbRefToken.Token)
	if RevokeErr != nil {
		respondwithJSON(w , 500 , "Something went revoking the refresh token")
		return 
	}

	respondwithJSON(w , 204 , nil)
}


func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter , req *http.Request) {
	type requestBody struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}

	var rqBody requestBody

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&rqBody)
	if err != nil {
		respondwithError(w , 401 , "Something went wrong")
		return
	}

	token , err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondwithError(w , 401 , "Invalid Header")
		return 
	}


	userID , err := auth.ValidateJWT(token , cfg.secretToken)
	if err != nil {
		respondwithError(w , 401 , "Error validating the token")
		return 
	}

	dbUser , err := cfg.db.UpdateUser(req.Context() , database.UpdateUserParams{
		Email : rqBody.Email,
		HashedPassword: auth.HashPassword(rqBody.Password),
		ID: userID,
	})
	if err != nil {
		respondwithError(w , 500 , "Error updating the user")
		return 
	}

	var user User 

	user.ID = dbUser.ID 
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	user.Email = dbUser.Email
	user.IsChirpyRed = dbUser.IsChirpyRed
	


	respondwithJSON(w , 200 , user)

}

func (cfg *apiConfig) handlerDelete(w http.ResponseWriter , req *http.Request) {

	token , err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondwithError(w , 401 , "Invalid Header")
		return 
	}


	userID , err := auth.ValidateJWT(token , cfg.secretToken)
	if err != nil {
		respondwithError(w , 401 , "Error validating the token")
		return 
	}

	pathParam := req.PathValue("chirpID")
	parsedUUID , err := uuid.Parse(pathParam)
	if err != nil {
		respondwithError(w , 500 , "Error Parsing the ID")
		return 
	}

	dbChirp , err := cfg.db.GetChirpByID(req.Context() , parsedUUID)
	if err != nil {
		respondwithError(w , 404 , "Not Found")
		return 
	}


	if userID != dbChirp.UserID{
		respondwithError(w , 403 , "Unauthorized Action")
		return 
	}

	deleteErr := cfg.db.DeleteChirp(req.Context() , dbChirp.ID)
	if deleteErr != nil {
		respondwithError(w , 500 , "error deleting the chirp")
		return 
	}

	respondwithJSON(w , 204 , nil)
}


func (cfg *apiConfig) handlerWebhook (w http.ResponseWriter , req *http.Request) {

	apiKey , err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondwithError(w , 401 , "Invalid header")
		return 
	}

	if apiKey != cfg.polkaKey {
		respondwithError(w , 401 , "Forbidden")
		return 
	}

	type requestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	var rqBody requestBody

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&rqBody)
	if err != nil {
		respondwithError(w , 401 , "Something went wrong")
		return
	}

	if rqBody.Event != "user.upgraded" {
		respondwithJSON(w , 204 , nil)
		return 
	}

	parsedUUID , err := uuid.Parse(rqBody.Data.UserID)
	if err != nil {
		respondwithError(w , 500 , "Error Parsing the ID")
		return 
	}
	_  , err = cfg.db.UpgradeMembership(req.Context() , parsedUUID)
	if err != nil {
		respondwithError(w , 404 , "Error Fetching the User")
	}

	respondwithJSON(w , 204 , nil)

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
	apiCfg.secretToken = os.Getenv("SECRET")
	apiCfg.polkaKey = os.Getenv("POLKA_KEY")



	serveMux.HandleFunc("POST /api/users" , func(w http.ResponseWriter , req *http.Request){
		type requestBody struct{
			Email string `json:"email"`
			Password string `json:"password"`
		}

		var rqBody requestBody

		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&rqBody)
		if err != nil {
			respondwithError(w , 500 , "Something went wrong")
			return
		}

		var user User 
		dbUser , err  := apiCfg.db.CreateUser(req.Context() ,  database.CreateUserParams{
			Email: rqBody.Email,
			HashedPassword: auth.HashPassword(rqBody.Password),
		})
		if err != nil {
			respondwithError(w , 400 , "Bad Input")
			return 
		}

		user.ID = dbUser.ID 
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.Email = dbUser.Email
		user.IsChirpyRed = dbUser.IsChirpyRed

		respondwithJSON(w , 201 , user)
	})

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

	serveMux.HandleFunc("GET /api/healthz" , func(w http.ResponseWriter , req *http.Request){
		w.Header().Set("Content-Type" , "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w , "OK")
	})

	log.Fatal(srv.ListenAndServe())
}




