package service

import (
	"context"
	"testing"

	"github.com/psds-microservice/helpy/db"
	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	conn, err := db.OpenInMemory()
	if err != nil {
		t.Fatalf("open in-memory: %v", err)
	}
	if err := conn.AutoMigrate(&model.User{}, &model.UserSession{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return conn
}

func TestUserAndAuth_CreateAndLogin(t *testing.T) {
	conn := testDB(t)
	userSvc := NewUserService(conn)
	authSvc := NewAuthService(conn)
	ctx := context.Background()

	req := &dto.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "secretpassword",
	}

	created, err := userSvc.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if created.ID == "" {
		t.Error("Expected ID to be set")
	}
	if created.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, created.Email)
	}

	loggedIn, err := authSvc.Login(ctx, "test@example.com", "secretpassword")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loggedIn.ID != created.ID {
		t.Errorf("Expected logged in user ID %s, got %s", created.ID, loggedIn.ID)
	}

	_, err = authSvc.Login(ctx, "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}
