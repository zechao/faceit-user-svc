package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zechao/faceit-user-svc/errors"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/user"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

var _ user.Repository = userRepository{}

func NewUserRepository(db *gorm.DB) user.Repository {
	return userRepository{db: db}
}

// CreateUser creates a new user record in the database. return nil if success.
func (r userRepository) CreateUser(ctx context.Context, u *user.User) (*user.User, error) {
	err := r.db.Create(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.ErrDuplicated
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return u, nil
}

// Delete perform soft delete operation, it won't delete the record from database but will
// set deleted_at field with a delete timestamp.
func (r userRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := r.db.Delete(&user.User{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Update perform update the given fields and return updated user.
func (r userRepository) UpdateUser(ctx context.Context, id uuid.UUID, fields map[string]any) (*user.User, error) {
	var u user.User
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&user.User{}).Where("id = ?", id).Updates(fields)
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
				return errors.ErrDuplicated
			}
			return fmt.Errorf("failed to update user: %w", res.Error)
		}

		if res.RowsAffected == 0 {
			return errors.ErrNotfound
		}
		err := tx.First(&u, "id = ?", id).Error
		if err != nil {
			return fmt.Errorf("failed to get user by ID: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// List list all users in the database. It should be able to filter and paginate the result based on the provided query object.
func (r userRepository) ListUsers(ctx context.Context, q query.Query) ([]user.User, error) {
	var users []user.User
	queryDB := q.ApplyQuery(r.db)
	err := queryDB.WithContext(ctx).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// Count returns the number of users in the database based on the provided filters.
func (r userRepository) CountUsers(ctx context.Context, filters map[string][]string) (int64, error) {
	db := r.db.WithContext(ctx).Model(&user.User{})
	for column, values := range filters {
		if len(values) > 0 {
			db = db.Where(column+" IN (?)", values)
		}
	}
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}
	return count, err
}

// GetUserByID implements user.Repository.
func (r userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrNotfound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &u, nil
}
