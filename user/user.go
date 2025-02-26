// Package user contains the user domain model and repository interface.
package user

//go:generate mockgen -source=user.go -destination=mocks/user_mock.go -package=mocks
import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zechao/faceit-user-svc/query"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
func (u *User) Update(input *UpdateUserInput) {
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
		u.Password = *input.Password
	}
	if input.Email != nil {
		u.Email = *input.Email
	}
	if input.Country != nil {
		u.Country = *input.Country
	}

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
	CreateUser(ctx context.Context, u *User) (*User, error)
	UpdateUser(ctx context.Context, input *UpdateUserInput) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, q query.Query) (*query.PaginationResponse[User], error)
}

// HashPassword hashes a password using bcrypt.
func HashPassword(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password")
	}
	return string(hash), nil
}

func ComparePassword(hash, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))

	return err == nil
}
