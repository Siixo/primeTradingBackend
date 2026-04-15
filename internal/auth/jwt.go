package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenClaims struct {
	// The ID of the user
	ID 				uint			`json:"id"`
	// The email address of the user
	Username 		string 			`json:"username"`
	// The role of the user
	Role 			string			`json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(secret []byte, id uint, username string, role string) (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &JWTTokenClaims{
		ID:       id,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func VerifyJWTToken(secret []byte, tokenString string) (*JWTTokenClaims, error) {
	claims := &JWTTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token)(interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenMalformed
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}
	
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func RefreshJWToken(secret []byte, claims *JWTTokenClaims) (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
