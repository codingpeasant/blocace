package main

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// CustomClaims is the customized claim based on the standard JWT claims
type CustomClaims struct {
	RoleName string `json:"roleName"`
	Address  string `json:"address"`
	jwt.StandardClaims
}

// issueToken authenticates the user and issue a jwt
func issueToken(address string, role Role) (string, error) {
	signingKey := []byte(secret)
	claims := CustomClaims{
		role.Name,
		address,
		jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Unix() + 600,
			Issuer:    "blocace",
			Audience:  "blocace user",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey)
	return tokenString, err
}

// verifyToken verifies is a jwt is valid
func verifyToken(tokenString string) (jwt.Claims, error) {
	signingKey := []byte(secret)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if token.Valid {
		return token.Claims, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("cannot validate the token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			return nil, errors.New("token expired")
		} else {
			return nil, errors.New("couldn't handle this token:" + err.Error())
		}
	} else {
		return nil, errors.New("couldn't handle this token:" + err.Error())
	}
}
