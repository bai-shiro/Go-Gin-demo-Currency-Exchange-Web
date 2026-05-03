package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error){
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 11)
	return string(hash), err
}

func GenerateJWT(username string) (string, error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username" : username,
		"exp" : time.Now().Add(time.Hour * 72).Unix(),
	})

	signedToken, err := token.SignedString([]byte("saki"))
	return "Bearer " + signedToken, err
}

func CheckPassword(pwd string, hashpwd string) bool{
	err := bcrypt.CompareHashAndPassword([]byte(hashpwd), []byte(pwd))
	return err == nil
}