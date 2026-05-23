// Package jwt provides JSON Web Token signing and verification for Nexgou modules.
//
// Usage:
//
//	var AppModule = nexgou.Module(nexgou.ModuleOptions{
//	    Imports:     []nexgou.IModule{nexgou.ConfigModule, nexgou.LogModule, jwt.Module},
//	    Controllers: []any{NewMyController},
//	})
//
//	func NewMyController(jwtSvc *jwt.JwtService) *MyController { ... }
package jwt

import (
	"errors"
	"strings"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/nexgou/server/src/common"
	"github.com/nexgou/server/src/config"
	"github.com/nexgou/server/src/logger"
)

// Claims embeds the standard JWT registered claims plus an arbitrary Data map.
type Claims struct {
	jwtlib.RegisteredClaims
	Data map[string]any `json:"data,omitempty"`
}

// JwtService signs and verifies HS256 JSON Web Tokens.
// Configuration is read from environment variables:
//
//	JWT_SECRET     — HMAC signing key (default: "nexgou-secret")
//	JWT_EXPIRATION — token TTL as a Go duration string (default: "24h")
//	JWT_ISSUER     — iss claim value (default: "nexgou")
type JwtService struct {
	secret []byte
	expiry time.Duration
	issuer string
	log    *logger.ScopedLogger
}

// NewJwtService creates a new JwtService.
// Depends on *config.ConfigService and *logger.LoggerService.
func NewJwtService(cfg *config.ConfigService, log *logger.LoggerService) *JwtService {
	secret := cfg.GetOrDefault("JWT_SECRET", "nexgou-secret")
	expStr := cfg.GetOrDefault("JWT_EXPIRATION", "24h")
	expiry, err := time.ParseDuration(expStr)
	if err != nil || expiry <= 0 {
		expiry = 24 * time.Hour
	}
	svc := &JwtService{
		secret: []byte(secret),
		expiry: expiry,
		issuer: cfg.GetOrDefault("JWT_ISSUER", "nexgou"),
		log:    log.WithContext("JwtService"),
	}
	svc.log.Info("initialized", "expiry", expiry, "issuer", svc.issuer)
	return svc
}

// Sign creates and signs a new JWT containing the given data payload.
// Returns the compact token string (header.payload.signature).
func (s *JwtService) Sign(data map[string]any) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(s.expiry)),
		},
		Data: data,
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// SignWithTTL creates a signed JWT with a custom expiration duration.
func (s *JwtService) SignWithTTL(data map[string]any, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
		},
		Data: data,
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// Verify parses and validates a compact token string.
// Returns the decoded Claims on success or an error if the token is invalid or expired.
func (s *JwtService) Verify(tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtlib.Token) (any, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, errors.New("nexgou/jwt: unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("nexgou/jwt: invalid token")
	}
	return claims, nil
}

// ExtractFromHeader reads the Bearer token from an Authorization header value.
// Returns an empty string if the header is missing or malformed.
func ExtractFromHeader(authHeader string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return ""
	}
	return strings.TrimPrefix(authHeader, prefix)
}

// ── JwtGuard ──────────────────────────────────────────────────────────────────

// JwtGuard is a route Guard that validates the Bearer JWT in the Authorization header.
// Attach it to any route that requires authentication:
//
//	nexgou.Get("/profile", c.Profile).Guard(&jwt.JwtGuard{Jwt: jwtSvc})
type JwtGuard struct {
	Jwt *JwtService
}

// CanActivate returns true when the request carries a valid, non-expired JWT.
func (g *JwtGuard) CanActivate(ctx *common.Context) (bool, error) {
	tokenStr := ExtractFromHeader(ctx.Header("Authorization"))
	if tokenStr == "" {
		return false, common.NewUnauthorizedException("missing or malformed Authorization header")
	}
	if _, err := g.Jwt.Verify(tokenStr); err != nil {
		return false, common.NewUnauthorizedException("invalid or expired token")
	}
	return true, nil
}
