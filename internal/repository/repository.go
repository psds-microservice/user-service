package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/psds-microservice/user-service/internal/dto"
	"github.com/psds-microservice/user-service/internal/model"
	"github.com/psds-microservice/user-service/pkg/constants"
)

type IUserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID) (*model.User, error)
	List(ctx context.Context, filters *dto.UserFilters) ([]*model.User, int64, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ListAvailableOperators(ctx context.Context, limit, offset int) ([]*model.User, int64, error)
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id.String()).Error
}

func (r *UserRepository) Get(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, filters *dto.UserFilters) ([]*model.User, int64, error) {
	var users []*model.User
	var count int64

	query := r.db.WithContext(ctx).Model(&model.User{})

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Role != "" {
		query = query.Where("role = ?", filters.Role)
	}
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", search, search)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err := query.Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListAvailableOperators(ctx context.Context, limit, offset int) ([]*model.User, int64, error) {
	var users []*model.User
	var count int64
	query := r.db.WithContext(ctx).Model(&model.User{}).
		Where("role = ? AND operator_status = ? AND is_available = ? AND is_active = ?",
			constants.RoleOperator, constants.OperatorStatusVerified, true, true)
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	query = query.Order("rating DESC NULLS LAST")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&users).Error
	return users, count, err
}
