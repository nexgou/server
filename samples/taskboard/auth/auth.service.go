package auth

import (
	"github.com/nexgou/server/src/logger"
	jwtmod "github.com/nexgou/server/src/module/jwt"
)

// AuthService handles credential verification and token issuance.
type AuthService struct {
	jwt *jwtmod.JwtService
	log *logger.ScopedLogger
}

// NewAuthService creates a new AuthService (injected by the IoC container).
func NewAuthService(jwt *jwtmod.JwtService, log *logger.LoggerService) *AuthService {
	return &AuthService{
		jwt: jwt,
		log: log.WithContext("AuthService"),
	}
}

// Login validates credentials and returns a signed JWT on success.
// For demo purposes only: accepts any username with password "password".
func (s *AuthService) Login(username, password string) (string, error) {
	if password != "password" {
		return "", nil
	}
	token, err := s.jwt.Sign(map[string]any{
		"sub":      username,
		"username": username,
	})
	if err != nil {
		s.log.Error("failed to sign token", "err", err)
		return "", err
	}
	s.log.Info("login successful", "username", username)
	return token, nil
}
