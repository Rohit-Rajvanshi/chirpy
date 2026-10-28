package auth


import (
	"net/http"
	"errors"
	"strings"
)

func GetBearerToken(headers http.Header) (string , error) {
	tokenString := headers.Get("Authorization")
	if tokenString == "" {
		return "" , errors.New("Error getting the TokenString")
	}

	stripToken := strings.Split(tokenString , " ")

	return stripToken[1] , nil 

}