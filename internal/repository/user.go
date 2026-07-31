package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name       string `gorm:"not null" validate:"required,min=2,max=100"`
	Email      string `gorm:"not null;uniqueIndex" validate:"required,email"`
	Password   Secret `gorm:"not null" validate:"required,min=8"`
	DisabledAt *time.Time
	TokenVersion int `gorm:"not null;default:0"`
}

func (r *Repository) CreateUser(
	ctx context.Context,
	name string,
	email string,
	password Secret,
) (*User, error) {
	user := &User{
		Name:     name,
		Email:    email,
		Password: password,
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

func (r *Repository) GetUser(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := r.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *User) error {
	return r.DB.WithContext(ctx).Save(user).Error
}
