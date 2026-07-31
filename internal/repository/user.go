package repository

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name     string `gorm:"not null" validate:"required,min=2,max=100"`
	Email    string `gorm:"not null;uniqueIndex" validate:"required,email"`
	Password Secret `gorm:"not null" validate:"required,min=8"`
}