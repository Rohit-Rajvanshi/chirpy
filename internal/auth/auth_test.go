package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidJWT (t *testing.T) {
	id := uuid.New()
	secret := "e234-423e"
	expiresIn := 10 * time.Second
	signedToken , err := MakeJWT(id , secret , expiresIn)
	if err != nil {
		t.Fatalf("Error making the JWT")
	}

	userId , err := ValidateJWT(signedToken , "e234-423e")
	if err != nil {
		t.Fatalf("Error validating")
	}

	if userId != id {
		t.Errorf("FAIL")
	}
}

func TestExpiryJWT (t *testing.T) {
	id := uuid.New()
	secret := "e234-423e"
	expiresIn := 10 * time.Second
	signedToken , err := MakeJWT(id , secret , expiresIn)
	if err != nil {
		t.Fatalf("Error making the JWT")
	}

	_ , err = ValidateJWT(signedToken , "e234-423e")
	if err != nil {
		t.Errorf("FAIL")
	}

}
