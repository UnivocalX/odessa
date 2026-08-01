package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name         string `gorm:"not null" validate:"required,min=2,max=100"`
	Email        string `gorm:"not null;uniqueIndex" validate:"required,email"`
	Password     Secret `gorm:"not null" validate:"required,min=8"`
	Role         string `gorm:"not null;default:user" validate:"required,oneof=user admin"`
	DisabledAt   *time.Time
	TokenVersion int `gorm:"not null;default:0"`
}

const (
	UserRole  = "user"
	AdminRole = "admin"
)

func (r *Repository) CreateUser(
	ctx context.Context,
	name string,
	email string,
	password Secret,
	role string,
) (*User, error) {
	user := &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
	}

	if err := validate.Struct(user); err != nil {
		return nil, err
	}

	if err := gorm.G[User](r.DB).Create(ctx, user); err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	if err := r.DB.WithContext(ctx).
		Where("email = ?", email).
		First(&user).
		Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmailIncludingDeleted is used by bootstrap logic because user
// deletion is implemented as a soft delete while email remains unique.
func (r *Repository) GetUserByEmailIncludingDeleted(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.DB.WithContext(ctx).Unscoped().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUser(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := r.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	if err := r.DB.WithContext(ctx).Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	return r.DB.WithContext(ctx).Save(user).Error
}

func (r *Repository) DeleteUser(ctx context.Context, id uint) error {
	result := r.DB.WithContext(ctx).Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) PurgeUser(ctx context.Context, id uint) error {
	result := r.DB.WithContext(ctx).Unscoped().Delete(&User{}, id)
	return result.Error
}
