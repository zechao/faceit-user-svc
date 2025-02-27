package http_test

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zechao/faceit-user-svc/errors"
	api "github.com/zechao/faceit-user-svc/http"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/user"
	"github.com/zechao/faceit-user-svc/user/mocks"
	"go.uber.org/mock/gomock"
)

var (
	//go:embed testdata/create_request.json
	createRequest []byte
	//go:embed testdata/update_request.json
	updateRequest []byte

	testUser = user.User{
		ID:        uuid.New(),
		FirstName: "John",
		LastName:  "Doe",
		NickName:  "AB123",
		Email:     "john.doe@example.com",
		Password:  "securepassword123",
		Country:   "ES",
	}
	errTest = errors.NewInternal("test error")
)

func TestValidateCreateUserRequest(t *testing.T) {

	t.Run("valid request", func(t *testing.T) {
		var req api.CreateUserRequest
		err := json.Unmarshal(createRequest, &req)
		assert.NoError(t, err)
		assert.NoError(t, req.Validate())
	})

	t.Run("invalid request", func(t *testing.T) {
		tests := []struct {
			name              string
			changeFunc        func(ss *api.CreateUserRequest)
			expectedErrDetail []errors.Detail
		}{
			{
				name: "all empty",
				changeFunc: func(r *api.CreateUserRequest) {
					r.Email = ""
					r.Password = ""
					r.FirstName = ""
					r.LastName = ""
					r.Country = ""
					r.NickName = ""
				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "first_name",
						Description: "first_name is required",
					},
					{
						Field:       "last_name",
						Description: "last_name is required",
					},
					{
						Field:       "nick_name",
						Description: "nick_name is required",
					},

					{
						Field:       "password",
						Description: "password is required",
					},
					{
						Field:       "email",
						Description: "email is required",
					},
					{
						Field:       "country",
						Description: "country must be a 2-letter ISO country code",
					},
				},
			},
			{
				name: "invalid password length",
				changeFunc: func(r *api.CreateUserRequest) {
					r.Password = "123456"

				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "password",
						Description: "password must between 8 and 40 characters long",
					},
				},
			},
			{
				name: "invalid email format",
				changeFunc: func(r *api.CreateUserRequest) {
					r.Email = "adas.com"

				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "email",
						Description: "invalid email format",
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var req api.CreateUserRequest
				err := json.Unmarshal(createRequest, &req)
				assert.NoError(t, err)
				tt.changeFunc(&req)
				err = req.Validate()

				wrongInputErr := new(errors.Error)
				errors.As(err, &wrongInputErr)
				assert.ElementsMatch(t, tt.expectedErrDetail, wrongInputErr.Details)
			})
		}
	})

}

func TestValidateUpdateUserRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		var req api.UpdateUserRequest
		err := json.Unmarshal(updateRequest, &req)
		assert.NoError(t, err)
		assert.NoError(t, req.Validate())
	})

	t.Run("invalid request", func(t *testing.T) {

		tests := []struct {
			name              string
			changeFunc        func(ss *api.UpdateUserRequest)
			expectedErrDetail []errors.Detail
		}{
			{
				name: "all empty",
				changeFunc: func(r *api.UpdateUserRequest) {
					emtyStr := ""
					r.Email = &emtyStr
					r.Password = &emtyStr
					r.FirstName = &emtyStr
					r.LastName = &emtyStr
					r.Country = &emtyStr
					r.NickName = &emtyStr
				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "first_name",
						Description: "first_name is required",
					},
					{
						Field:       "last_name",
						Description: "last_name is required",
					},
					{
						Field:       "nick_name",
						Description: "nick_name is required",
					},

					{
						Field:       "password",
						Description: "password is required",
					},
					{
						Field:       "email",
						Description: "email is required",
					},
					{
						Field:       "country",
						Description: "country must be a 2-letter ISO country code",
					},
				},
			},
			{
				name: "invalid password length",
				changeFunc: func(r *api.UpdateUserRequest) {
					shortPassword := "123456"
					r.Password = &shortPassword

				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "password",
						Description: "password must between 8 and 40 characters long",
					},
				},
			},
			{
				name: "invalid email format",
				changeFunc: func(r *api.UpdateUserRequest) {
					wrongEmail := "adas.com"
					r.Email = &wrongEmail

				},
				expectedErrDetail: []errors.Detail{
					{
						Field:       "email",
						Description: "invalid email format",
					},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var req api.UpdateUserRequest
				err := json.Unmarshal(updateRequest, &req)
				assert.NoError(t, err)
				tt.changeFunc(&req)
				err = req.Validate()

				wrongInputErr := new(errors.Error)
				errors.As(err, &wrongInputErr)
				assert.ElementsMatch(t, tt.expectedErrDetail, wrongInputErr.Details)
			})
		}
	})
}

func setupRouter() *gin.Engine {
	// without log
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestCreateUser(t *testing.T) {
	router := setupRouter()
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockService(ctrl)
	handler := api.NewUserHandler(mockService)
	handler.RegisterRoutes(router)

	var validReq api.CreateUserRequest
	err := json.Unmarshal(createRequest, &validReq)
	assert.NoError(t, err)

	tests := map[string]struct {
		requestBody    io.Reader
		mockSetup      func()
		expectedStatus int
	}{
		"fail by payload error": {
			requestBody:    strings.NewReader("invalid payload"),
			expectedStatus: http.StatusBadRequest,
		},
		"fail by validation error": {
			requestBody: func() io.Reader {
				invalidReq := validReq
				invalidReq.Country = "invalid"
				data, err := json.Marshal(invalidReq)
				assert.NoError(t, err)
				return strings.NewReader(string(data))
			}(),
			expectedStatus: http.StatusBadRequest,
		},
		"fail by service error": {
			requestBody: strings.NewReader(string(createRequest)),
			mockSetup: func() {
				mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, errTest)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		"success valid request": {
			requestBody: strings.NewReader(string(createRequest)),
			mockSetup: func() {
				mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&testUser, nil)
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request, err = http.NewRequest(http.MethodPost, "/users", tt.requestBody)
			assert.NoError(t, err)

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			router.ServeHTTP(recorder, ctx.Request)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	router := setupRouter()
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockService(ctrl)
	handler := api.NewUserHandler(mockService)
	handler.RegisterRoutes(router)

	var validReq api.UpdateUserRequest
	err := json.Unmarshal(createRequest, &validReq)
	assert.NoError(t, err)

	tests := map[string]struct {
		id             string
		requestBody    io.Reader
		mockSetup      func()
		expectedStatus int
	}{
		"fail by wrong id": {
			id:             "wrong id",
			requestBody:    nil,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		"fail by payload error": {
			id:             testUser.ID.String(),
			requestBody:    strings.NewReader("invalid payload"),
			expectedStatus: http.StatusBadRequest,
		},
		"fail by validation error": {
			id: testUser.ID.String(),
			requestBody: func() io.Reader {
				invalidReq := validReq
				emptyText := ""
				invalidReq.Country = &emptyText
				data, err := json.Marshal(invalidReq)
				assert.NoError(t, err)
				return strings.NewReader(string(data))
			}(),
			expectedStatus: http.StatusBadRequest,
		},
		"fail by service error": {
			id:          testUser.ID.String(),
			requestBody: strings.NewReader(string(createRequest)),
			mockSetup: func() {
				mockService.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil, errTest)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		"success valid request": {
			id:          testUser.ID.String(),
			requestBody: strings.NewReader(string(createRequest)),
			mockSetup: func() {
				mockService.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(&testUser, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request, err = http.NewRequest(http.MethodPatch, "/users/"+tt.id, tt.requestBody)
			assert.NoError(t, err)

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			router.ServeHTTP(recorder, ctx.Request)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	router := setupRouter()
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockService(ctrl)
	handler := api.NewUserHandler(mockService)
	handler.RegisterRoutes(router)

	var validReq api.UpdateUserRequest
	err := json.Unmarshal(createRequest, &validReq)
	assert.NoError(t, err)

	tests := map[string]struct {
		id             string
		requestBody    io.Reader
		mockSetup      func()
		expectedStatus int
	}{
		"fail by wrong id": {
			id:             "wrong id",
			requestBody:    nil,
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		"fail by service error": {
			id:          testUser.ID.String(),
			requestBody: strings.NewReader(string(createRequest)),
			mockSetup: func() {
				mockService.EXPECT().DeleteUser(gomock.Any(), testUser.ID).Return(errTest)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		"success valid request": {
			id:          testUser.ID.String(),
			requestBody: strings.NewReader(string(createRequest)),
			mockSetup: func() {
				mockService.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request, err = http.NewRequest(http.MethodDelete, "/users/"+tt.id, tt.requestBody)
			assert.NoError(t, err)

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			router.ServeHTTP(recorder, ctx.Request)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestListUser(t *testing.T) {
	router := setupRouter()
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockService(ctrl)
	handler := api.NewUserHandler(mockService)
	handler.RegisterRoutes(router)

	tests := map[string]struct {
		params         string
		requestBody    io.Reader
		mockSetup      func()
		expectedStatus int
	}{
		"fail by query error empty paramSortBy": {
			params:         "page=0&page_size=0",
			mockSetup:      nil,
			expectedStatus: http.StatusBadRequest,
		},
		"fail by service": {
			params: "page=1&page_size=10",
			mockSetup: func() {
				mockService.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(nil, errTest)
			},
			expectedStatus: http.StatusInternalServerError,
		},

		"success ": {
			params: "page=1&page_size=10",
			mockSetup: func() {
				mockService.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(&query.PaginationResponse[user.User]{
					Page:         1,
					PageSize:     10,
					TotalRecords: 1,
					SortBy:       "id",
					SortOrder:    "asc",
					Filters:      map[string][]string{},
					Data:         []user.User{testUser},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var err error
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request, err = http.NewRequest(http.MethodGet, "/users?"+tt.params, nil)
			assert.NoError(t, err)

			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			router.ServeHTTP(recorder, ctx.Request)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}
