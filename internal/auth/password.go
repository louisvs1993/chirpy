package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error){
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
    	fmt.Println("Error:", err)
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password string, hash string) (bool, error){
	correctPassword, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
    	fmt.Println("Error:", err)
		return false, err
	}
	return correctPassword, nil
}