package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword ...
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash ...
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ToJWKS converts a RSA public key to a JSON encoded jwk set
func ToJWKS(pub *rsa.PublicKey) ([]byte, error) {
	// See https://github.com/golang/crypto/blob/master/acme/jws.go#L90
	// https://tools.ietf.org/html/rfc7518#section-6.3.1
	n := pub.N
	e := big.NewInt(int64(pub.E))
	// Field order is important.
	// See https://tools.ietf.org/html/rfc7638#section-3.3 for details.
	return []byte(fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`,
		base64.RawURLEncoding.EncodeToString(e.Bytes()),
		base64.RawURLEncoding.EncodeToString(n.Bytes()),
	)), nil
}

// ToPEM converts a RSA private key to PEM format
func ToPEM(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
}

// SignJwt ...
func (a *Authenticator) SignJwt(claims Claims) (string, error) {
	expirationTime := time.Now().Add(time.Duration(a.ExpireSeconds) * time.Second)
	// Add metadata to the the JWT claims, which should already include a user ID
	sc := claims.GetStandardClaims()
	// In JWT, the expiry time is expressed as unix milliseconds
	sc.ExpiresAt = expirationTime.Unix()
	sc.Issuer = a.Issuer
	sc.Audience = a.Audience

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "0"
	token.Header["alg"] = "RS256"

	return token.SignedString(a.SignKey)
}
