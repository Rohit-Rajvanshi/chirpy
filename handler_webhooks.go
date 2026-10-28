package main

import (
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/Rohit-Rajvanshi/chirpy/internal/auth"
)


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
