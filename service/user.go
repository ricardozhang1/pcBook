package service

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// User contains user's information
type User struct {
	UserName       string
	HashedPassword string
	Role           string
}

// NewUser return a new user
func NewUser(username, password, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %v", err)
	}
	user := &User{
		UserName: username,
		HashedPassword: string(hashedPassword),
		Role: role,
	}
	return user, nil
}

// IsCorrectPassword check the password provided is correct or not
func (user *User) IsCorrectPassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return false
	}
	return true
}

// Clone returns a clone this user
func (user *User) Clone() *User {
	return &User{
		UserName: user.UserName,
		HashedPassword: user.HashedPassword,
		Role: user.Role,
	}
}

