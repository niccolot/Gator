package auth

import (
	"context"
	"fmt"

	"github.com/niccolot/BlogAggregator/internal/database"
	"github.com/niccolot/BlogAggregator/internal/state"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

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

func AskNewPassword() (string, error) {
	fmt.Println("Choose a password: ")
	pass, err := term.ReadPassword(0)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}

	fmt.Println("Repeat password: ")
	pass2, err := term.ReadPassword(0)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}

	if string(pass) != string(pass2) {
		return "", fmt.Errorf("enter the same password")
	}

	hashed_password, err := HashPassword(string(pass))
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}

	return hashed_password, nil
}

func AskPassword(user *database.User) error {
	fmt.Println("Insert password: ")
	pass, err := term.ReadPassword(0)
	if err != nil {
		return fmt.Errorf("failed to read password: %v", err)
	}

	errCheck := CheckPasswordHash(string(pass), user.HashedPassword)
	if errCheck != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}

func CheckSuperUser(s *state.State, user *database.User) error {
	if user.ID != s.Cfg.SuperUserID {
		return fmt.Errorf("you must be superuser to run this command")
	}

	errPass := AskPassword(user)
	if errPass != nil {
		return errPass
	}

	return nil
}

func ChangePassword(user *database.User, s *state.State) error {
	hashed_password, err := AskNewPassword()
	if err != nil {
		return err
	}

	pars := &database.ChangePasswordParams{ID: user.ID, HashedPassword: hashed_password}
	err = s.Db.ChangePassword(context.Background(), *pars)
	if err != nil {
		return fmt.Errorf("failed to change password: %v", err)
	}

	return nil
}