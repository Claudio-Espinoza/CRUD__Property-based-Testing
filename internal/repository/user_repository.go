package repository

import (
	"sync"

	"property-based/internal/domain"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetAll() ([]*domain.User, error)
	Update(user *domain.User) error
	Delete(id string) error
	Count() int
}

type InMemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[string]*domain.User
	emails map[string]string
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[string]*domain.User),
		emails: make(map[string]string),
	}
}

func (r *InMemoryUserRepository) Create(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return domain.ErrAlreadyExists
	}
	if _, exists := r.emails[user.Email]; exists {
		return domain.ErrAlreadyExists
	}
	r.users[user.ID] = user.Clone()
	r.emails[user.Email] = user.ID

	return nil
}

func (r *InMemoryUserRepository) GetByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return user.Clone(), nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userID, exists := r.emails[email]
	if !exists {
		return nil, domain.ErrNotFound
	}

	user := r.users[userID]
	return user.Clone(), nil
}

func (r *InMemoryUserRepository) GetAll() ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user.Clone())
	}

	return users, nil
}

func (r *InMemoryUserRepository) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldUser, exists := r.users[user.ID]
	if !exists {
		return domain.ErrNotFound
	}

	if oldUser.Email != user.Email {
		if existingUserID, exists := r.emails[user.Email]; exists && existingUserID != user.ID {
			return domain.ErrAlreadyExists
		}

		delete(r.emails, oldUser.Email)
		r.emails[user.Email] = user.ID
	}

	r.users[user.ID] = user.Clone()
	return nil
}

func (r *InMemoryUserRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return domain.ErrNotFound
	}

	delete(r.users, id)
	delete(r.emails, user.Email)

	return nil
}

func (r *InMemoryUserRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.users)
}
