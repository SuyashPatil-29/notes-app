package models

import (
	"time"
)

type User struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name"`
	Email     string     `json:"email" gorm:"index:idx_email,unique"`
	ImageUrl  *string    `json:"imageUrl"`
	Notebooks []Notebook `json:"notebooks,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type AuthenticatedUser struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	ImageUrl *string `json:"imageUrl"`
}

// TableName tells GORM to use the 'users' table for AuthenticatedUser
func (AuthenticatedUser) TableName() string {
	return "users"
}
