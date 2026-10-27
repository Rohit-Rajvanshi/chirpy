package auth

import (
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)





func ValidateJWT(tokenString , tokenSecret string) (uuid.UUID , error) {
	claims := &jwt.RegisteredClaims{}

	_ , err := jwt.ParseWithClaims(tokenString , claims , func(token *jwt.Token) (any , error){
		return []byte(tokenSecret) , nil 
	})
	if err != nil {
		return uuid.Nil , err 
	}

	id , err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil , err 
	}

	return id , nil 
}