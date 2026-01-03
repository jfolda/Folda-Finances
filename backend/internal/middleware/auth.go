package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const ClaimsKey contextKey = "claims"

// JWKS structures
type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type AuthMiddleware struct {
	jwtSecret   string
	supabaseURL string
	publicKeys  map[string]*ecdsa.PublicKey
	keysMutex   sync.RWMutex
	lastFetch   time.Time
}

func NewAuthMiddleware(jwtSecret string, supabaseURL string) *AuthMiddleware {
	am := &AuthMiddleware{
		jwtSecret:   jwtSecret,
		supabaseURL: supabaseURL,
		publicKeys:  make(map[string]*ecdsa.PublicKey),
	}

	// Fetch keys on initialization
	if supabaseURL != "" {
		if err := am.fetchJWKS(); err != nil {
			log.Printf("Warning: Failed to fetch JWKS: %v", err)
		}
	} else {
		log.Printf("⚠️  WARNING: SUPABASE_URL is not set - ES256 JWT validation will fail!")
	}

	return am
}

func (am *AuthMiddleware) fetchJWKS() error {
	jwksURL := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", strings.TrimSuffix(am.supabaseURL, "/"))

	resp, err := http.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	am.keysMutex.Lock()
	defer am.keysMutex.Unlock()

	for _, key := range jwks.Keys {
		if key.Kty == "EC" && key.Crv == "P-256" {
			pubKey, err := am.parseECPublicKey(key)
			if err != nil {
				log.Printf("Warning: Failed to parse key %s: %v", key.Kid, err)
				continue
			}
			am.publicKeys[key.Kid] = pubKey
		}
	}

	am.lastFetch = time.Now()
	log.Printf("✓ Fetched %d public keys from Supabase JWKS", len(am.publicKeys))
	return nil
}

func (am *AuthMiddleware) parseECPublicKey(key JWK) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y: %w", err)
	}

	pubKey := &ecdsa.PublicKey{
		Curve: nil, // Will be set based on Crv
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	// Set curve based on Crv parameter
	switch key.Crv {
	case "P-256":
		pubKey.Curve = elliptic.P256()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", key.Crv)
	}

	return pubKey, nil
}

func (am *AuthMiddleware) getPublicKey(kid string) (*ecdsa.PublicKey, error) {
	am.keysMutex.RLock()
	key, exists := am.publicKeys[kid]
	am.keysMutex.RUnlock()

	if exists {
		return key, nil
	}

	// Refresh keys if not found or stale (>1 hour)
	if time.Since(am.lastFetch) > time.Hour {
		if err := am.fetchJWKS(); err != nil {
			return nil, err
		}

		am.keysMutex.RLock()
		key, exists = am.publicKeys[kid]
		am.keysMutex.RUnlock()

		if exists {
			return key, nil
		}
	}

	return nil, fmt.Errorf("public key not found for kid: %s", kid)
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

		userID, claims, err := am.validateToken(tokenString)
		if err != nil {
			// Log the actual error for debugging
			log.Printf("JWT validation error: %v", err)
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (am *AuthMiddleware) validateToken(tokenString string) (uuid.UUID, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		alg := token.Method.Alg()
		log.Printf("Token signing method: %s", alg)

		switch alg {
		case "ES256":
			// Get kid (key ID) from token header
			kid, ok := token.Header["kid"].(string)
			if !ok {
				log.Printf("Token missing kid header. Headers: %+v", token.Header)
				return nil, errors.New("token missing kid header")
			}
			log.Printf("Looking up public key for kid: %s", kid)

			// Get public key from JWKS
			pubKey, err := am.getPublicKey(kid)
			if err != nil {
				log.Printf("Failed to get public key for kid %s: %v", kid, err)
				return nil, fmt.Errorf("failed to get public key: %w", err)
			}
			log.Printf("✓ Found public key for kid: %s", kid)
			return pubKey, nil
		case "HS256":
			// Use secret for HS256
			log.Printf("Using HS256 with JWT secret")
			return []byte(am.jwtSecret), nil
		default:
			log.Printf("Unexpected signing method: %s", alg)
			return nil, errors.New("unexpected signing method: " + alg)
		}
	})

	if err != nil {
		log.Printf("JWT parse error: %v", err)
		return uuid.Nil, nil, err
	}

	if !token.Valid {
		log.Printf("Token marked as invalid after parsing")
		return uuid.Nil, nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Failed to cast token claims to MapClaims")
		return uuid.Nil, nil, errors.New("invalid token claims")
	}

	log.Printf("Token claims: %+v", claims)

	// Extract user ID from Supabase token (usually in 'sub' claim)
	sub, ok := claims["sub"].(string)
	if !ok {
		log.Printf("Missing or invalid 'sub' claim in token")
		return uuid.Nil, nil, errors.New("missing sub claim")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		log.Printf("Invalid UUID in sub claim: %s", sub)
		return uuid.Nil, nil, errors.New("invalid user ID in token")
	}

	log.Printf("✓ Successfully validated token for user: %s", userID)
	return userID, claims, nil
}

func GetUserID(r *http.Request) (uuid.UUID, error) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}

func GetUserEmail(r *http.Request) (string, error) {
	claims, ok := r.Context().Value(ClaimsKey).(jwt.MapClaims)
	if !ok {
		return "", errors.New("claims not found in context")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", errors.New("email not found in claims")
	}
	return email, nil
}
