package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenAccess  = "access"
	TokenRefresh = "refresh"
)

// Claims are the JWT claims shared by every service. Never rename "id".
type Claims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Type  string `json:"typ"`
	jwt.RegisteredClaims
}

// IssueToken signs claims with HS256 then wraps the JWT with AES-256-GCM.
func IssueToken(cfg AppConfig, claims Claims, ttl time.Duration) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(ttl))
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", err
	}
	return encrypt(cfg.EncryptionKey, []byte(signed))
}

// ParseToken unwraps, verifies the signature and checks expiry.
func ParseToken(cfg AppConfig, token string) (*Claims, error) {
	raw, err := decrypt(cfg.EncryptionKey, token)
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(string(raw), claims, func(*jwt.Token) (any, error) {
		return []byte(cfg.SecretKey), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// BearerToken extracts the token from an "Authorization: Bearer <token>" header.
func BearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", false
	}
	return token, true
}

func gcm(key string) (cipher.AEAD, error) {
	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func encrypt(key string, plain []byte) (string, error) {
	aead, err := gcm(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(aead.Seal(nonce, nonce, plain, nil)), nil
}

func decrypt(key, encoded string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	aead, err := gcm(key)
	if err != nil {
		return nil, err
	}
	if len(data) < aead.NonceSize() {
		return nil, errors.New("invalid token")
	}
	plain, err := aead.Open(nil, data[:aead.NonceSize()], data[aead.NonceSize():], nil)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	return plain, nil
}
