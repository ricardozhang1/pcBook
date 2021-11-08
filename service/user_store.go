package service

import (
	"fmt"
	"sync"
)

// UserStore is an interface to store users
type UserStore interface {
	// Save save the user to the store
	Save(user *User) error
	// Find finds a user by username
	Find(username string) (*User, error)
}

// InMemoryUserStore stores users in memory
type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

// NewInMemoryUserStore returns a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

// Save save the user to the store
func (store *InMemoryUserStore) Save(user *User) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user1 := store.users[user.UserName]
	if user1 != nil {
		return ErrAlreadyExists
	}
	store.users[user.UserName] = user.Clone()
	return nil
}

// Find finds a user by username
func (store *InMemoryUserStore) Find(username string) (*User, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	user := store.users[username]
	if user == nil {
		return nil, fmt.Errorf("user %v is not exist", username)
	}
	return user.Clone(), nil
}

