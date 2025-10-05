package domain

import (
	"regexp"
	"strings"
	"time"
)

type User struct {
	ID        string
	Name      string
	Email     string
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var nameRegex = regexp.MustCompile(`^[A-Za-z][A-Za-z ]{0,48}[A-Za-z]$`)

func (u *User) Validate() error {
	u.Name = strings.TrimSpace(u.Name)

	if len(u.Name) < 2 || len(u.Name) > 50 {
		return ErrInvalidUserName
	}
	if !nameRegex.MatchString(u.Name) {
		return ErrInvalidUserName
	}

	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidUserEmail
	}

	if u.Age < 1 || u.Age > 150 {
		return ErrInvalidUserAge
	}

	return nil
}

func NewUser(id, name, email string, age int) (*User, error) {
	now := time.Now().UTC()
	user := &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Age:       age,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) Clone() *User {
	return &User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Age:       u.Age,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
