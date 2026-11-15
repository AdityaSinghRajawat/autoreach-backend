package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

type accessTokenClaims struct {
	Roles       []string `json:"roles"`
	CandidateID string   `json:"cid,omitempty"`
	jwt.RegisteredClaims
}

func IssueAccessToken(
	principal sharedTypes.Principal,
	issuedAt time.Time,
) (string, time.Time, error) {
	expiresAt := issuedAt.Add(config.GetAccessTokenTTL())
	claims := accessTokenClaims{
		Roles:       principal.Roles,
		CandidateID: principal.CandidateID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.GetTokenIssuer(),
			Subject:   principal.Subject,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString(config.GetJWTSecret())
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

func ParseAccessToken(raw string) (sharedTypes.Principal, error) {
	claims := accessTokenClaims{}

	_, err := jwt.ParseWithClaims(
		raw,
		&claims,
		func(_ *jwt.Token) (any, error) { return config.GetJWTSecret(), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(config.GetTokenIssuer()),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return sharedTypes.Principal{}, fmt.Errorf("%w: %w", config.ErrInvalidToken, err)
	}

	if claims.Subject == "" || len(claims.Roles) == 0 {
		return sharedTypes.Principal{}, fmt.Errorf("%w: incomplete claims", config.ErrInvalidToken)
	}

	return sharedTypes.NewUserPrincipal(claims.Subject, claims.CandidateID, claims.Roles), nil
}

func NewOpaqueToken() (string, string, error) {
	raw := make([]byte, config.GetRefreshTokenBytes())
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate opaque token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(raw)

	return token, Fingerprint(token), nil
}

func SetPrincipal(c *gin.Context, principal sharedTypes.Principal) {
	c.Set(config.GetPrincipalContextKey(), principal)
}

func GetPrincipal(c *gin.Context) (sharedTypes.Principal, bool) {
	value, found := c.Get(config.GetPrincipalContextKey())
	if !found {
		return sharedTypes.Principal{}, false
	}

	principal, isPrincipal := value.(sharedTypes.Principal)

	return principal, isPrincipal
}
