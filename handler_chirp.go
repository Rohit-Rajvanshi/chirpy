package main

import (
	"net/http"
	"encoding/json"
	"github.com/Rohit-Rajvanshi/chirpy/internal/database"
	"github.com/Rohit-Rajvanshi/chirpy/internal/auth"
	"github.com/google/uuid"
	"sort"
)

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