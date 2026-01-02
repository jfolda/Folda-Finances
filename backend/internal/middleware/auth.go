package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	jwtSecret    string
	jwtPublicKey *ecdsa.PublicKey
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:    jwtSecret,
		jwtPublicKey: nil, // Will be set separately if needed
	}
}

func (am *AuthMiddleware) SetPublicKey(publicKeyPEM string) error {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	ecdsaKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("not an ECDSA public key")
	}

	am.jwtPublicKey = ecdsaKey
	return nil
}

func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		userID, err := am.validateToken(tokenString)
		if err != nil {
			// Log the actual error for debugging
			println("JWT validation error:", err.Error())
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (am *AuthMiddleware) validateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		switch token.Method.Alg() {
		case "ES256":
			// Use ECDSA public key for ES256
			if am.jwtPublicKey == nil {
				return nil, errors.New("public key not configured for ES256")
			}
			return am.jwtPublicKey, nil
		case "HS256":
			// Use secret for HS256
			return []byte(am.jwtSecret), nil
		default:
			println("Unexpected signing method:", token.Method.Alg())
			return nil, errors.New("unexpected signing method: " + token.Method.Alg())
		}
	})

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}

	// Extract user ID from Supabase token (usually in 'sub' claim)
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("missing sub claim")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	return userID, nil
}

func GetUserID(r *http.Request) (uuid.UUID, error) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}
