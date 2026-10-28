package main

import (
	"net/http"
	"encoding/json"
	"github.com/Rohit-Rajvanshi/chirpy/internal/database"
	"github.com/Rohit-Rajvanshi/chirpy/internal/auth"
	"time"
)

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
