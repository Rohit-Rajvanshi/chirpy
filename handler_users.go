package main 

import (
	"net/http"
	"encoding/json"
	"github.com/Rohit-Rajvanshi/chirpy/internal/auth"
	"github.com/Rohit-Rajvanshi/chirpy/internal/database"

)


func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter , req *http.Request){
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
	dbUser , err  := cfg.db.CreateUser(req.Context() ,  database.CreateUserParams{
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

