package auth

import (
	"github.com/alexedwards/argon2id"
	"log"
)


func HashPassword(password string) string {

	hash , err := argon2id.CreateHash(password , argon2id.DefaultParams)
	if err != nil {
		log.Fatal(err)
	}

	return hash 
}