package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zechao/faceit-user-svc/errors"
	"github.com/zechao/faceit-user-svc/postgres"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/user"
	"gorm.io/gorm"
)

var testUser = user.User{
	ID:        uuid.New(),
	FirstName: "jin",
	LastName:  "zechao",
	NickName:  "zen",
	Email:     "zechao@gmail.com",
	Password:  "superpassword",
	Country:   "ES",
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDatabase(t)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	t.Run("valid user", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser
		err := repo.CreateUser(ctx, &tu)
		assert.NoError(t, err)
		assert.NotZero(t, tu.CreatedAt, "timestamp should be set")
		assert.NotZero(t, tu.UpdatedAt, "timestamp should be set")
	})

	t.Run("fail by duplicated email", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser
		err = repo.CreateUser(ctx, &tu)
		assert.NoError(t, err)
		err = repo.CreateUser(ctx, &tu)
		assert.ErrorIs(t, err, errors.ErrDuplicated,
			"should return error because of duplicated email")
	})

	t.Run("fail by invalid input", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		err = repo.CreateUser(ctx, &user.User{
			Country: "wrong",
		})
		assert.Error(t, err, "should return error, because country only accept 2 char")
	})
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDatabase(t)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	t.Run("success update", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser
		err := tx.Create(&tu).Error
		assert.NoError(t, err)
		tu.Country = "UK"
		tu.NickName = "nickname"

		err = repo.UpdateUser(ctx, &tu)
		assert.NoError(t, err)
	})

	t.Run("success create if not exist", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser
		err := repo.UpdateUser(ctx, &tu)
		assert.NoError(t, err)
		assert.NotZero(t, tu.CreatedAt, "should create with timestamp")
		assert.NotZero(t, tu.UpdatedAt, "should create with timestamp")

	})

	t.Run("fail updating to existing email", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser

		err := tx.Create(&tu).Error
		assert.NoError(t, err)
		tu2 := testUser
		tu2.ID = uuid.New()
		tu2.Email = "existing@example.com"
		err = tx.Create(&tu2).Error
		assert.NoError(t, err)

		tu.Email = tu2.Email // Update to existing email
		err = repo.UpdateUser(ctx, &tu)
		assert.ErrorIs(t, err, errors.ErrDuplicated)
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDatabase(t)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	t.Run("success delete", func(t *testing.T) {
		repo := postgres.NewUserRepository(db)
		tu := testUser

		err := db.Create(&tu).Error
		assert.NoError(t, err)
		err = repo.DeleteUser(ctx, tu.ID)
		assert.NoError(t, err)
		// shouldn't return any
		err = db.First(&tu, tu.ID).Error
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		// should still exist in database with deleted at set
		var u user.User
		err = db.Unscoped().Where("ID = ?", tu.ID).Find(&u).Error
		assert.NoError(t, err)
		assert.Equal(t, u.ID, tu.ID)
	})

	t.Run("success nothing deleted", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)

		err = repo.DeleteUser(ctx, uuid.New())
		assert.NoError(t, err)
	})

}

func TestListUser(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDatabase(t)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	t.Run("success list empty users", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize: 1000,
		})
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("success list all users", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize: 1000,
		})
		assert.NoError(t, err)
		assert.Len(t, res, len(users))
	})

	t.Run("success list with page", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize: 3,
			Page:     2,
		})
		assert.NoError(t, err)
		assert.Len(t, res, 3)
	})
	t.Run("success empty list out of page", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize: 5,
			Page:     3,
		})
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("success filter by country with page", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES", "GB")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize: 6,
			Page:     2,
			Filters:  map[string][]string{"country": {"ES"}},
		})
		assert.NoError(t, err)
		assert.Len(t, res, 4)
	})

	t.Run("success filter by multiple filter with page", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES", "GB")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize: 10,
			Page:     1,
			Filters: map[string][]string{
				"country":    {"ES", "GB"},
				"first_name": {"zechao0"},
				"last_name":  {"jin0"},
				"nick_name":  {"zen0"},
				"password":   {"superpassword"},
			},
		})
		assert.NoError(t, err)
		assert.Len(t, res, 2)
	})

	t.Run("success list filter by multiple filter with order", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES", "GB")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize:  20,
			Page:      1,
			SortOrder: "asc",
			SortBy:    "first_name",
			Filters: map[string][]string{
				"country": {"ES", "GB"},
			},
		})
		assert.NoError(t, err)
		assert.Len(t, res, 20)
		assert.Equal(t, res[0].FirstName, "zechao0")
		assert.Equal(t, res[len(users)-1].FirstName, "zechao9")
	})

	t.Run("success list filter by id default sort", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES", "GB")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.ListUsers(ctx, query.Query{
			PageSize: 10,
			Page:     1,
			Filters: map[string][]string{
				"id": {users[0].ID.String(), users[1].ID.String()},
			},
		})
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, users[1].ID, res[1].ID)
		assert.Equal(t, users[0].ID, res[0].ID)
	})

}

func createUsers(db *gorm.DB, n int, countries ...string) ([]user.User, error) {
	users := make([]user.User, 0, n)
	for i := 0; i < n; i++ {
		for _, country := range countries {
			users = append(users, user.User{
				ID:        uuid.New(),
				FirstName: fmt.Sprintf("zechao%d", i),
				LastName:  fmt.Sprintf("jin%d", i),
				NickName:  fmt.Sprintf("zen%d", i),
				Email:     fmt.Sprintf("user%d.%s@example.com", i, country),
				Password:  "superpassword",
				Country:   country,
			})
		}

	}
	err := db.Create(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
