package service

import (
	"context"
	"testing"
	"time"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
	"gorm.io/gorm"
)

// StubUserRepository is a simple in-memory repository for testing
type StubUserRepository struct {
	users  map[uint]*model.User
	nextID uint
}

func NewStubUserRepository() *StubUserRepository {
	return &StubUserRepository{
		users:  make(map[uint]*model.User),
		nextID: 1,
	}
}

func (r *StubUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	user.ID = r.nextID
	r.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return user, nil
}

func (r *StubUserRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if _, exists := r.users[user.ID]; !exists {
		return nil, gorm.ErrRecordNotFound
	}
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return user, nil
}

func (r *StubUserRepository) Delete(ctx context.Context, id uint) error {
	delete(r.users, id)
	return nil
}

func (r *StubUserRepository) Get(ctx context.Context, id uint) (*model.User, error) {
	if user, exists := r.users[id]; exists {
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
	return nil, nil // Not found
}

func TestUserService_CreateAndLogin(t *testing.T) {
	repo := NewStubUserRepository()
	s := NewUserService(repo)
	ctx := context.Background()

	// 1. Create User
	req := &dto.CreateUserRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "secretpassword",
		Notes:    "Some notes",
	}

	created, err := s.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if created.Id == 0 {
		t.Error("Expected ID to be set")
	}
	if created.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, created.Email)
	}

	// 2. Login
	loggedIn, err := s.Login(ctx, "test@example.com", "secretpassword")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loggedIn.Id != created.Id {
		t.Errorf("Expected logged in user ID %d, got %d", created.Id, loggedIn.Id)
	}

	// 3. Login with wrong password
	_, err = s.Login(ctx, "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}
