package auth

import (
	"time"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)


func MakeJWT(userID uuid.UUID ,tokenSecret string , expiresIn time.Duration) (string , error) {
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256 , jwt.RegisteredClaims{
		Issuer : "chirpy-access" ,
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject: userID.String(),
	})

	signedToken , err := newToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "" , err
	}

	return signedToken , nil 
}

