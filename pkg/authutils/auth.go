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

func VerifyToken(tokenString string) error {
	cfg := env.Get()
	secretKey := []byte(cfg.SecretKey)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func GetDataJWT(tokenString string) (any, error) {
	cfg := env.Get()
	secretKey := []byte(cfg.SecretKey)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid map claims")
	}

	return claims["data"], nil
}
