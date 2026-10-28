package auth

import(
	"net/http"
	"errors"
	"strings"
)

func GetAPIKey(headers http.Header) (string , error){
	apiKey := headers.Get("Authorization")
	if apiKey == "" {
		return "" , errors.New("Error getting the TokenString")
	}

	stripaApiKey := strings.Split(apiKey , " ")

	return stripaApiKey[1] , nil 
}