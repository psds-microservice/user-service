package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/internal/repository"
)

type IUserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	GetUser(ctx context.Context, id uint) (*dto.UserResponse, error)
	ListUsers(ctx context.Context, filters *dto.UserFilters) ([]*dto.UserResponse, int64, error)
	Login(ctx context.Context, email, password string) (*dto.UserResponse, error)
}

type UserService struct {
	repo repository.IUserRepository
}

func NewUserService(repo repository.IUserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:     req.Email,
		Name:      req.Name,
		Password:  hashedPassword,
		Notes:     req.Notes,
		CreatedBy: req.CreatedBy,
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return toUserResponse(createdUser), nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	user.Email = req.Email
	user.Name = req.Name
	user.Notes = req.Notes
	user.UpdatedBy = req.UpdatedBy

	if req.Password != "" {
		hashed, err := hashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashed
	}

	updatedUser, err := s.repo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return toUserResponse(updatedUser), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) GetUser(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return toUserResponse(user), nil
}

func (s *UserService) ListUsers(ctx context.Context, filters *dto.UserFilters) ([]*dto.UserResponse, int64, error) {
	users, count, err := s.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	var responses []*dto.UserResponse
	for _, u := range users {
		responses = append(responses, toUserResponse(u))
	}

	return responses, count, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if !checkPassword(user.Password, password) {
		return nil, errors.New("invalid credentials")
	}

	return toUserResponse(user), nil
}

// Helpers

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func toUserResponse(u *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Notes:     u.Notes,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
