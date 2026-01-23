package domain

import "time"

type User struct {
	ID           string    `json:"id" dynamodbav:"UserId"`
	Email        string    `json:"email" dynamodbav:"Email"`
	Name         string    `json:"name" dynamodbav:"Name"`
	PasswordHash string    `json:"-" dynamodbav:"PasswordHash"`
	CreatedAt    time.Time `json:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt    time.Time `json:"updatedAt" dynamodbav:"UpdatedAt"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
