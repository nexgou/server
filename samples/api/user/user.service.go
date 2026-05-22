package user

import (
	"github.com/nexgou/server/src/config"
	"github.com/nexgou/server/src/logger"
)

// UserService handles all business logic for the User resource.
type UserService struct {
	cfg *config.ConfigService
	log *logger.ScopedLogger
}

// NewUserService creates a new UserService (used by the IoC container).
// ConfigService and LoggerService are injected automatically.
func NewUserService(cfg *config.ConfigService, log *logger.LoggerService) *UserService {
	return &UserService{
		cfg: cfg,
		log: log.WithContext("UserService"),
	}
}

// FindAll returns a list of all users.
func (s *UserService) FindAll() []map[string]string {
	s.log.Info("FindAll called")
	return []map[string]string{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
		{"id": "3", "name": "Charlie"},
	}
}

// FindOne returns a single user by ID.
func (s *UserService) FindOne(id string) map[string]string {
	s.log.Info("FindOne called", "id", id)
	return map[string]string{"id": id, "name": "User " + id}
}

// Create persists a new user and returns it.
func (s *UserService) Create(name string) map[string]string {
	s.log.Info("Create called", "name", name)
	return map[string]string{"id": "new-id", "name": name}
}
