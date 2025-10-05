package service

import (
	"time"

	"github.com/google/uuid"

	"property-based/internal/domain"
	"property-based/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(name, email string, age int) (*domain.User, error) {
	id := uuid.New().String()
	user, err := domain.NewUser(id, name, email, age)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(id string) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) GetUserByEmail(email string) (*domain.User, error) {
	return s.repo.GetByEmail(email)
}

func (s *UserService) GetAllUsers() ([]*domain.User, error) {
	return s.repo.GetAll()
}

func (s *UserService) UpdateUser(id, name, email string, age int) (*domain.User, error) {
	existingUser, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	updatedUser := &domain.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Age:       age,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}

	if err := updatedUser.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(updatedUser); err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *UserService) DeleteUser(id string) error {
	return s.repo.Delete(id)
}

func (s *UserService) CountUsers() int {
	return s.repo.Count()
}
