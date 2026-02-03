package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/pkg/constants"
)

// StubUserRepository — in-memory репозиторий для тестов (UUID).
type StubUserRepository struct {
	users map[string]*model.User
}

func NewStubUserRepository() *StubUserRepository {
	return &StubUserRepository{
		users: make(map[string]*model.User),
	}
}

func (r *StubUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	r.users[user.ID] = user
	return user, nil
}

func (r *StubUserRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if _, exists := r.users[user.ID]; !exists {
		return nil, nil
	}
	r.users[user.ID] = user
	return user, nil
}

func (r *StubUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id.String())
	return nil
}

func (r *StubUserRepository) Get(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if user, exists := r.users[id.String()]; exists {
		return user, nil
	}
	return nil, nil
}

func (r *StubUserRepository) List(ctx context.Context, filters *dto.UserFilters) ([]*model.User, int64, error) {
	var list []*model.User
	for _, u := range r.users {
		list = append(list, u)
	}
	return list, int64(len(list)), nil
}

func (r *StubUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (r *StubUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (r *StubUserRepository) ListAvailableOperators(ctx context.Context, limit, offset int) ([]*model.User, int64, error) {
	var list []*model.User
	for _, u := range r.users {
		if u.Role == constants.RoleOperator && u.OperatorStatus == constants.OperatorStatusVerified && u.IsAvailable {
			list = append(list, u)
		}
	}
	return list, int64(len(list)), nil
}

// StubSessionRepository для тестов.
type StubSessionRepository struct{}

func (StubSessionRepository) Create(ctx context.Context, s *model.UserSession) (*model.UserSession, error) {
	return s, nil
}
func (StubSessionRepository) Update(ctx context.Context, s *model.UserSession) (*model.UserSession, error) {
	return s, nil
}
func (StubSessionRepository) Get(ctx context.Context, id uuid.UUID) (*model.UserSession, error) {
	return nil, nil
}
func (StubSessionRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.UserSession, int64, error) {
	return nil, 0, nil
}
func (StubSessionRepository) CountActiveByUserID(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}
func (StubSessionRepository) ListActiveByUserID(ctx context.Context, userID string) ([]*model.UserSession, error) {
	return nil, nil
}

func TestUserService_CreateAndLogin(t *testing.T) {
	repo := NewStubUserRepository()
	s := NewUserService(repo, StubSessionRepository{})
	ctx := context.Background()

	req := &dto.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "secretpassword",
	}

	created, err := s.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if created.ID == "" {
		t.Error("Expected ID to be set")
	}
	if created.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, created.Email)
	}
	if created.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, created.Username)
	}

	loggedIn, err := s.Login(ctx, "test@example.com", "secretpassword")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loggedIn.ID != created.ID {
		t.Errorf("Expected logged in user ID %s, got %s", created.ID, loggedIn.ID)
	}

	_, err = s.Login(ctx, "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}
