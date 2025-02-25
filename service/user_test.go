package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/service"
	"github.com/zechao/faceit-user-svc/user"
	"github.com/zechao/faceit-user-svc/user/mocks"
	"go.uber.org/mock/gomock"
)

var (
	tesUser = user.User{
		FirstName: "jin",
		LastName:  "zechao",
		NickName:  "zen",
		Email:     "zechao@gmail.com",
		Password:  "superpassword",
		Country:   "ES",
	}
	errTest = errors.New("test error")
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	t.Run("should create user successfully", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)

		reqUser := tesUser

		expectedUser := reqUser
		expectedUser.ID = uuid.New()

		// compare that input user is the same as the user created in the repository
		// the password should be hashed, so we compare using ComparePassword
		mockUserRepo.EXPECT().CreateUser(ctx, gomock.Cond(func(uu *user.User) bool {
			expectedUser.Password = uu.Password
			return uu.FirstName == reqUser.FirstName &&
				uu.LastName == reqUser.LastName &&
				uu.NickName == reqUser.NickName &&
				uu.Email == reqUser.Email &&
				uu.Country == reqUser.Country &&
				user.ComparePassword(uu.Password, reqUser.Password)

		})).Return(&expectedUser, nil)

		res, err := svc.CreateUser(ctx, &tesUser)

		assert.Equal(t, &expectedUser, res)
		assert.Nil(t, err)
	})

	t.Run("should fail when repository return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)

		mockUserRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(nil, errTest)

		res, err := svc.CreateUser(ctx, &tesUser)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})

}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	t.Run("should update user successfully", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)
		expectedUser := user.User{
			ID:        uuid.New(),
			FirstName: "John",
			LastName:  "Doe",
			NickName:  "johndoe",
			Email:     "johndoe@example.com",
			Country:   "USA",
			Password:  "newpassword",
		}

		reqInput := map[string]any{
			"first_name": expectedUser.FirstName,
			"last_name":  expectedUser.LastName,
			"nick_name":  expectedUser.NickName,
			"email":      expectedUser.NickName,
			"password":   expectedUser.Password,
			"country":    expectedUser.NickName,
		}

		mockUserRepo.EXPECT().UpdateUser(ctx, expectedUser.ID, gomock.Cond(func(filter map[string]any) bool {
			for k, v := range filter {
				if k == "password" && !user.ComparePassword(v.(string), expectedUser.Password) {
					return false
				}
				if reqInput[k] != v {
					return false
				}
			}
			return true

		})).Return(&expectedUser, nil)

		res, err := svc.UpdateUser(ctx, expectedUser.ID, reqInput)

		assert.Equal(t, &expectedUser, res)
		assert.Nil(t, err)
	})

	t.Run("should return error when repository returns error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)
		mockUserRepo.EXPECT().UpdateUser(ctx, gomock.Any(), gomock.Any()).Return(nil, errTest)

		res, err := svc.UpdateUser(ctx, uuid.New(), map[string]any{})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	t.Run("should delete user successfully", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)

		id := uuid.New()
		mockUserRepo.EXPECT().DeleteUser(ctx, id).Return(nil)

		err := svc.DeleteUser(ctx, id)

		assert.Nil(t, err)
	})

	t.Run("fail deleting user", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)

		id := uuid.New()
		mockUserRepo.EXPECT().DeleteUser(ctx, id).Return(errTest)

		err := svc.DeleteUser(ctx, id)

		assert.ErrorIs(t, err, errTest)
	})
}

func TestListUsers(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	t.Run("success list user one page", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)

		q := query.Query{
			Page:      1,
			PageSize:  10,
			SortOrder: "asc",
			SortBy:    "created_by",
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
		}

		users := []user.User{
			tesUser,
			tesUser,
		}

		pageResult := &query.PaginationResponse[user.User]{
			Page:         1,
			PageSize:     10,
			SortOrder:    "asc",
			SortBy:       "created_by",
			TotalRecords: 2,
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
			Data: users,
		}

		mockUserRepo.EXPECT().CountUsers(ctx, q.Filters).Return(int64(len(users)), nil)
		mockUserRepo.EXPECT().ListUsers(ctx, q).Return(users, nil)

		res, err := svc.ListUsers(ctx, q)

		assert.Equal(t, pageResult, res)
		assert.NoError(t, err)
	})

	t.Run("success list last page", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)
		users := []user.User{
			tesUser,
			tesUser,
		}
		totalCount := int64(12) // Assuming there are 12 users in total
		q := query.Query{
			Page:      2,
			PageSize:  10,
			SortOrder: "asc",
			SortBy:    "created_by",
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
		}

		// Adjust the query to simulate the last page, total count is 12, so last page has 2 elements
		pageResult := &query.PaginationResponse[user.User]{
			Page:         2,
			PageSize:     10,
			SortOrder:    "asc",
			SortBy:       "created_by",
			TotalRecords: totalCount,
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
			Data: users,
		}

		// Adjust the query to simulate the last page, total count is 12, so last page has 2 elements
		mockUserRepo.EXPECT().CountUsers(ctx, q.Filters).Return(totalCount, nil)
		mockUserRepo.EXPECT().ListUsers(ctx, q).Return(users, nil)
		res, err := svc.ListUsers(ctx, q)

		assert.Equal(t, pageResult, res)
		assert.NoError(t, err)
	})

	t.Run("empty list when total count is zero", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)

		q := query.Query{
			Page:      1,
			PageSize:  10,
			SortOrder: "asc",
			SortBy:    "created_by",
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
		}

		pageResult := &query.PaginationResponse[user.User]{
			Page:         1,
			PageSize:     10,
			SortOrder:    "asc",
			SortBy:       "created_by",
			TotalRecords: 0,
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
			Data: []user.User{},
		}

		mockUserRepo.EXPECT().CountUsers(ctx, q.Filters).Return(int64(0), nil)

		res, err := svc.ListUsers(ctx, q)

		assert.Equal(t, pageResult, res)
		assert.NoError(t, err)
	})

	t.Run("empty list when page requested is beyond the last page", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)
		totalCount := int64(12) // Assuming there are 12 users in total
		q := query.Query{
			Page:      3,
			PageSize:  10,
			SortOrder: "asc",
			SortBy:    "created_by",
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
		}

		// Adjust the query to simulate the last page, total count is 12, so last page has 2 elements
		pageResult := &query.PaginationResponse[user.User]{
			Page:         3,
			PageSize:     10,
			SortOrder:    "asc",
			SortBy:       "created_by",
			TotalRecords: totalCount,
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
			Data: []user.User{},
		}

		// page request is beyond the last page, so we expect an empty list
		mockUserRepo.EXPECT().CountUsers(ctx, q.Filters).Return(totalCount, nil)
		res, err := svc.ListUsers(ctx, q)

		assert.Equal(t, pageResult, res)
		assert.NoError(t, err)
	})

	t.Run("fail when count return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)
		q := query.Query{
			Page:      1,
			PageSize:  10,
			SortOrder: "asc",
			SortBy:    "created_by",
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
		}
		// page request is beyond the last page, so we expect an empty list
		mockUserRepo.EXPECT().CountUsers(ctx, q.Filters).Return(int64(0), errTest)
		res, err := svc.ListUsers(ctx, q)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("fail when list user return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		svc := service.NewUserService(mockUserRepo)
		q := query.Query{
			Page:      1,
			PageSize:  10,
			SortOrder: "asc",
			SortBy:    "created_by",
			Filters: map[string][]string{
				"name": {"John Doe", "Jane Doe2"},
			},
		}
		// page request is beyond the last page, so we expect an empty list
		mockUserRepo.EXPECT().CountUsers(ctx, q.Filters).Return(int64(15), nil)
		mockUserRepo.EXPECT().ListUsers(ctx, q).Return(nil, errTest)
		res, err := svc.ListUsers(ctx, q)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})

}
