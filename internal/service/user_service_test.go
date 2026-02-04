package service

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
)

// testDB создаёт in-memory SQLite БД с миграциями для тестов.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserSession{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestUserService_CreateAndLogin(t *testing.T) {
	db := testDB(t)
	s := NewUserService(db)
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
