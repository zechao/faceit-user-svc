// Package user contains the user domain model and repository interface.
package user

//go:generate mockgen -source=user.go -destination=mocks/user_mock.go -package=mocks
import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zechao/faceit-user-svc/query"

	"github.com/zechao/faceit-user-svc/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrPasswordTooLong is an error returned when the password is too long to hash.
var ErrPasswordTooLong = errors.NewWrongInput("can't hash password")

// EventType represents the type of event related to a user.
type EventType string

const (
	UserCreated EventType = "UserCreated"
	UserUpdated EventType = "UserUpdated"
	UserDeleted EventType = "UserDeleted"
)

// User represents a user domain model.
type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	NickName  string
	Email     string
	Password  string
	Country   string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// CreateUserInput represents the input for creating a user.
type CreateUserInput struct {
	FirstName string
	LastName  string
	NickName  string
	Password  string
	Email     string
	Country   string
}

// NewUser creates a new user with the provided input. It hashes the password before storing it.
// it also generates a new UUID for the user.
func NewUser(input *CreateUserInput) (*User, error) {
	hashedPassword, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	u := User{
		ID:        uuid.New(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		NickName:  input.NickName,
		Password:  hashedPassword,
		Email:     input.Email,
		Country:   input.Country,
	}
	return &u, nil
}

type UpdateUserInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	NickName  *string
	Password  *string
	Email     *string
	Country   *string
}

// Update updates the user fields with the provided input.
func (u *User) Update(input *UpdateUserInput) error {

	if input.FirstName != nil {
		u.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		u.LastName = *input.LastName
	}
	if input.NickName != nil {
		u.NickName = *input.NickName
	}
	if input.Password != nil {
		hashedPassword, err := HashPassword(*input.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		u.Password = hashedPassword
	}
	if input.Email != nil {
		u.Email = *input.Email
	}
	if input.Country != nil {
		u.Country = *input.Country
	}
	return nil
}

// TableName returns the table name for the user model.
func (User) TableName() string {
	return "user_svc.users"
}

// Repository defines the interface for user data access operations.
type Repository interface {
	CreateUser(ctx context.Context, u *User) (*User, error)
	UpdateUser(ctx context.Context, u *User) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, q query.Query) ([]User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	CountUsers(ctx context.Context, filters map[string][]string) (int64, error)
}

// Service defines the interface for user business logic operations.
type Service interface {
	CreateUser(ctx context.Context, u *CreateUserInput) (*User, error)
	UpdateUser(ctx context.Context, input *UpdateUserInput) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, q query.Query) (*query.PaginationResponse[User], error)
}

// HashPassword hashes a password using bcrypt.
func HashPassword(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrPasswordTooLong
	}
	return string(hash), nil
}

// ComparePassword compares a hashed password with a plain text password.
func ComparePassword(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))

	return err == nil
}
