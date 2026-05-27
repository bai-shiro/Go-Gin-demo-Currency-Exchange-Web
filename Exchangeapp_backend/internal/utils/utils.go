package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 11)
	return string(hash), err
}

func CheckPassword(pwd string, hashpwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashpwd), []byte(pwd))
	return err == nil
}
