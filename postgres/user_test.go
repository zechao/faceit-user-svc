package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
		res, err := repo.CreateUser(ctx, &tu)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, res.ID, "id should be set")
		assert.NotZero(t, res.CreatedAt, "timestamp should be set")
		assert.NotZero(t, res.UpdatedAt, "timestamp should be set")
	})

	t.Run("fail by duplicated email", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser
		res, err := repo.CreateUser(ctx, &tu)
		assert.NotNil(t, res)
		assert.NoError(t, err)
		res, err = repo.CreateUser(ctx, &tu)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errors.ErrDuplicated,
			"should return error because of duplicated email")
	})

	t.Run("fail by invalid input", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		res, err := repo.CreateUser(ctx, &user.User{
			Country: "wrong",
		})
		assert.Error(t, err, "should return error, because country only accept 2 char")
		assert.Nil(t, res)
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

		res, err := repo.UpdateUser(ctx, tu.ID, map[string]any{
			"country":   tu.Country,
			"nick_name": tu.NickName,
		})
		assert.NoError(t, err)
		assertEqualUser(t, tu, res)
	})

	t.Run("success create if not exist", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser
		res, err := repo.UpdateUser(ctx, tu.ID, map[string]interface{}{
			"country": "UK",
		})
		assert.ErrorIs(t, err, errors.ErrNotfound)
		assert.Nil(t, res)
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

		res, err := repo.UpdateUser(ctx, tu.ID, map[string]any{
			"email": tu2.Email, // Update to existing email
		})
		assert.ErrorIs(t, err, errors.ErrDuplicated)
		assert.Nil(t, res)
	})
	t.Run("update success", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		tu := testUser

		err := tx.Create(&tu).Error
		assert.NoError(t, err)

		expected := testUser
		expected.LastName = "newLastName"
		expected.FirstName = "newFirstName"
		expected.NickName = "newNickName"
		expected.Email = "newEmail@example.com"
		expected.Country = "ES"
		expected.Password = "newPassword"

		res, err := repo.UpdateUser(ctx, tu.ID, map[string]any{
			"first_name": expected.FirstName,
			"last_name":  expected.LastName,
			"nick_name":  expected.NickName,
			"email":      expected.Email,
			"country":    expected.Country,
			"password":   expected.Password,
		})

		assert.NoError(t, err)
		assert.Equal(t, expected.FirstName, res.FirstName)
		assert.Equal(t, expected.LastName, res.LastName)
		assert.Equal(t, expected.NickName, res.NickName)
		assert.Equal(t, expected.Email, res.Email)
		assert.Equal(t, expected.Country, res.Country)
		assert.Equal(t, expected.Password, res.Password)
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

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDatabase(t)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	t.Run("success get existing user by id", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()

		repo := postgres.NewUserRepository(tx)
		tu := testUser
		err := tx.Create(&tu).Error
		assert.NoError(t, err)

		res, err := repo.GetUserByID(ctx, tu.ID)
		assert.NoError(t, err)
		assertEqualUser(t, tu, res)
	})
	t.Run("fail by not found user by id", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()

		repo := postgres.NewUserRepository(tx)

		res, err := repo.GetUserByID(ctx, uuid.New())
		assert.ErrorIs(t, err, errors.ErrNotfound)
		assert.Nil(t, res)
	})
}

func TestCountUsers(t *testing.T) {
	ctx := context.Background()
	db, err := setupTestDatabase(t)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	t.Run("success count empty users", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		res, err := repo.CountUsers(ctx, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, int64(0), res)
	})

	t.Run("success count all users", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES", "GB")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.CountUsers(ctx, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, len(users), res)
	})

	t.Run("success count with filter users", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES", "GB")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.CountUsers(ctx, map[string][]string{
			"country": {"ES"},
		})
		assert.NoError(t, err)
		assert.EqualValues(t, 10, res)
	})

	t.Run("success count with multiplefilter users", func(t *testing.T) {
		tx := db.Begin()
		defer tx.Rollback()
		repo := postgres.NewUserRepository(tx)
		users, err := createUsers(tx, 10, "ES", "GB", "US")
		assert.NoError(t, err)
		assert.NotNil(t, users)
		res, err := repo.CountUsers(ctx, map[string][]string{
			"country":    {"ES", "GB"},
			"first_name": {"zechao0"},
		})
		assert.NoError(t, err)
		assert.EqualValues(t, 2, res)
	})
}
func assertEqualUser(t *testing.T, expected user.User, actual *user.User) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.FirstName, actual.FirstName)
	assert.Equal(t, expected.LastName, actual.LastName)
	assert.Equal(t, expected.NickName, actual.NickName)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.Password, actual.Password)
	assert.Equal(t, expected.Country, actual.Country)
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Millisecond)

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
