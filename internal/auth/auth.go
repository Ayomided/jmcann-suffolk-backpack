package auth

import (
	"context"
	"fmt"
	"time"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	contextKeyUserRole contextKey = "userRole"
	contextKeyUserID   contextKey = "userID"
)

func CreateAccessToken(user *model.User, secret string, expiry int) (accessToken string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        user.ID,
		"exp":        time.Now().Add(time.Minute * time.Duration(expiry)).Unix(),
		"iat":        time.Now().Unix(),
		"iss":        "iji",
		"loggedInAs": user.Role,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func CreateRefreshToken(user *model.User, secret string, expiry int) (refreshToken string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * time.Duration(expiry)).Unix(),
	})

	refereshTokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return refereshTokenString, nil
}

func ExtractIDFromToken(requestToken, secret string) (string, error) {
	token, err := jwt.Parse(requestToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return "", fmt.Errorf("Invalid Token")
	}

	id, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("sub claim missing or not a string")
	}

	return id, nil
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(contextKeyUserID).(string)
	return id, ok
}

func UserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(contextKeyUserRole).(string)
	return role, ok
}

func ExtractClaimsFromToken(requestToken, secret string) (userID, role string, err error) {
	token, err := jwt.Parse(requestToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	userID, ok = claims["sub"].(string)
	if !ok {
		return "", "", fmt.Errorf("sub claim missing or not a string")
	}

	role, ok = claims["loggedInAs"].(string)
	if !ok {
		return "", "", fmt.Errorf("loggedInAs claim missing or not a string")
	}

	return userID, role, nil
}
