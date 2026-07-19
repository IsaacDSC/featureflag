package authutils

import (
	"errors"
	"fmt"
	"github.com/IsaacDSC/featureflag/internal/env"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func CreateToken(data any) (string, error) {
	cfg := env.Get()
	secretKey := []byte(cfg.SecretKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"data": data,
			"exp":  time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyingKeyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return []byte(env.Get().SecretKey), nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, verifyingKeyFunc)

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func GetDataJWT(tokenString string) (any, error) {
	token, err := jwt.Parse(tokenString, verifyingKeyFunc)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid map claims")
	}

	return claims["data"], nil
}
