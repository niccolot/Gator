package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	password_bytes := []byte(password)
	hash, errHashing := bcrypt.GenerateFromPassword(password_bytes, bcrypt.DefaultCost)
	if errHashing != nil {
		return "", errHashing
	}

	return string(hash), nil
} 

func CheckPasswordHash(password string, hash string) error {
	errCompPass := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errCompPass != nil {
		return errCompPass
	}

	return nil
}