package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockevent "github.com/zechao/faceit-user-svc/event/mocks"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/service"
	"github.com/zechao/faceit-user-svc/user"
	"github.com/zechao/faceit-user-svc/user/mocks"
	"go.uber.org/mock/gomock"
)

var (
	tesUser = user.User{
		FirstName: "john",
		LastName:  "doe",
		NickName:  "j.d",
		Email:     "j.d@gmail.com",
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
		eventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, eventHandler)

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
		eventHandler.EXPECT().SendEvent(ctx, string(user.UserCreated), expectedUser.ID).
			Return(nil)

		res, err := svc.CreateUser(ctx, &tesUser)

		assert.Equal(t, &expectedUser, res)
		assert.Nil(t, err)
	})

	t.Run("should fail when repository return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		eventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, eventHandler)

		mockUserRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(nil, errTest)

		res, err := svc.CreateUser(ctx, &tesUser)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})
	t.Run("should fail when event handler return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		eventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, eventHandler)

		expectedUser := tesUser
		expectedUser.ID = uuid.New()

		mockUserRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(&expectedUser, nil)
		eventHandler.EXPECT().SendEvent(ctx, string(user.UserCreated), expectedUser.ID).
			Return(errTest)

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
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)
		currentUser := tesUser
		currentUser.ID = uuid.New()

		expectedUser := user.User{
			FirstName: "zechao",
			LastName:  "jin",
			NickName:  "zen",
			Email:     "zechao@gmail.com",
			Password:  "superhasedpassword",
			Country:   "GB",
		}
		updateInput := user.UpdateUserInput{
			ID:        expectedUser.ID,
			FirstName: &expectedUser.FirstName,
			LastName:  &expectedUser.LastName,
			NickName:  &expectedUser.NickName,
			Email:     &expectedUser.Email,
			Country:   &expectedUser.Country,
			Password:  &expectedUser.Password,
		}

		mockUserRepo.EXPECT().GetUserByID(ctx, expectedUser.ID).Return(&currentUser, nil)
		mockUserRepo.EXPECT().UpdateUser(ctx, gomock.Cond(func(uu *user.User) bool {
			return uu.FirstName == expectedUser.FirstName &&
				uu.LastName == expectedUser.LastName &&
				uu.NickName == expectedUser.NickName &&
				uu.Email == expectedUser.Email &&
				uu.Country == expectedUser.Country &&
				// check that the password is hashed
				user.ComparePassword(uu.Password, expectedUser.Password)

		})).Return(&expectedUser, nil)
		mockEventHandler.EXPECT().SendEvent(ctx, string(user.UserUpdated), expectedUser.ID).
			Return(nil)
		res, err := svc.UpdateUser(ctx, &updateInput)

		assert.Equal(t, &expectedUser, res)
		assert.Nil(t, err)
	})

	t.Run("should return error when get return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)
		testID := uuid.New()
		mockUserRepo.EXPECT().GetUserByID(ctx, testID).Return(nil, errTest)

		res, err := svc.UpdateUser(ctx, &user.UpdateUserInput{
			ID: testID,
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("should return error when uppdate return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)
		currentUser := tesUser
		mockUserRepo.EXPECT().GetUserByID(ctx, gomock.Any()).Return(&currentUser, nil)
		mockUserRepo.EXPECT().UpdateUser(ctx, gomock.Any()).Return(nil, errTest)

		res, err := svc.UpdateUser(ctx, &user.UpdateUserInput{
			ID: uuid.New(),
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})

	t.Run("should return error when event handler return error", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)
		currentUser := tesUser
		mockUserRepo.EXPECT().GetUserByID(ctx, gomock.Any()).Return(&currentUser, nil)
		mockUserRepo.EXPECT().UpdateUser(ctx, gomock.Any()).Return(&currentUser, nil)
		mockEventHandler.EXPECT().SendEvent(ctx, string(user.UserUpdated), currentUser.ID).
			Return(errTest)
		res, err := svc.UpdateUser(ctx, &user.UpdateUserInput{
			ID: uuid.New(),
		})

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errTest)
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	id := uuid.New()
	tests := map[string]struct {
		setupMocks  func(mockUserRepo *mocks.MockRepository, mockEventHandler *mockevent.MockEventHandler)
		expectedErr error
	}{
		"should delete user successfully": {
			setupMocks: func(mockUserRepo *mocks.MockRepository, mockEventHandler *mockevent.MockEventHandler) {
				mockUserRepo.EXPECT().DeleteUser(ctx, id).Return(nil)
				mockEventHandler.EXPECT().SendEvent(ctx, string(user.UserDeleted), id).Return(nil)
			},
			expectedErr: nil,
		},
		"fail deleting user": {
			setupMocks: func(mockUserRepo *mocks.MockRepository, mockEventHandler *mockevent.MockEventHandler) {
				mockUserRepo.EXPECT().DeleteUser(ctx, id).Return(errTest)
			},
			expectedErr: errTest,
		},
		"fail sending event": {
			setupMocks: func(mockUserRepo *mocks.MockRepository, mockEventHandler *mockevent.MockEventHandler) {
				mockUserRepo.EXPECT().DeleteUser(ctx, id).Return(nil)
				mockEventHandler.EXPECT().SendEvent(ctx, string(user.UserDeleted), id).Return(errTest)
			},
			expectedErr: errTest,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockRepository(ctrl)
			mockEventHandler := mockevent.NewMockEventHandler(ctrl)
			svc := service.NewUserService(mockUserRepo, mockEventHandler)

			tc.setupMocks(mockUserRepo, mockEventHandler)

			err := svc.DeleteUser(ctx, id)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			}

		})
	}
}

func TestListUsers(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)

	t.Run("success list user one page", func(t *testing.T) {
		mockUserRepo := mocks.NewMockRepository(ctrl)
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)

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
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)
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
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)

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
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)

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
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)

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
		mockEventHandler := mockevent.NewMockEventHandler(ctrl)
		svc := service.NewUserService(mockUserRepo, mockEventHandler)

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
