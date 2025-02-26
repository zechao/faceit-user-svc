package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zechao/faceit-user-svc/errors"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/user"
)

var (
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	ErrorInvalidUserID = errors.NewWrongInput("invalid user id")
)

type UserHandler struct {
	service user.Service
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/users", h.CreateUser)
	router.PATCH("/users/:id", h.UpdateUser)
	router.GET("/users", h.ListUsers)
	router.DELETE("/users/:id", h.DeleteUser)
}

func NewUserHandler(service user.Service) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// CreateUser handles the creation of a new user and returns the created user details without password.
// usually returning http status code 201 Created is enough to indicate that the resource was created successfully.
// but in this case we are returning the created user details as well. which is also common practice.
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var req CreateUserRequest

	err := json.NewDecoder(ctx.Request.Body).Decode(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errors.ErrInvalidPayload)
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := h.service.CreateUser(ctx.Request.Context(), &user.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		NickName:  req.NickName,
		Password:  req.Password,
		Email:     req.Email,
		Country:   req.Country,
	})
	if err != nil {
		handlerError(ctx, err)
		return
	}
	res := &UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		NickName:  user.NickName,
		Email:     user.Email,
		Country:   user.Country,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	ctx.JSON(http.StatusCreated, res)
}

func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errors.NewWrongInput("invalid user id"))
		return
	}

	var req UpdateUserRequest
	err = json.NewDecoder(ctx.Request.Body).Decode(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errors.ErrInvalidPayload)
		return
	}

	if err := req.Vaildate(); err != nil {
		handlerError(ctx, err)
		return
	}

	user, err := h.service.UpdateUser(ctx.Request.Context(), &user.UpdateUserInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		NickName:  req.NickName,
		Country:   req.Country,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		handlerError(ctx, err)
		return
	}

	res := &UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		NickName:  user.NickName,
		Email:     user.Email,
		Country:   user.Country,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *UserHandler) ListUsers(ctx *gin.Context) {

	queryInput, err := query.QueryFromURL(ctx.Request.URL.Query())
	if err != nil {
		handlerError(ctx, err)
	}

	listRes, err := h.service.ListUsers(ctx.Request.Context(), *queryInput)
	if err != nil {
		handlerError(ctx, err)
		return
	}

	res := ListUsersResponse{
		Page:         listRes.Page,
		PageSize:     listRes.PageSize,
		TotalRecords: listRes.TotalRecords,
		SortBy:       listRes.SortBy,
		SortOrder:    listRes.SortOrder,
		Filters:      listRes.Filters,
		Users:        make([]UserResponse, len(listRes.Data)),
	}
	for i, u := range listRes.Data {
		res.Users[i] = UserResponse{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			NickName:  u.NickName,
			Email:     u.Email,
			Country:   u.Country,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	ctx.JSON(http.StatusOK, res)

}

func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorInvalidUserID)
		return
	}

	err = h.service.DeleteUser(ctx.Request.Context(), id)
	if err != nil {
		handlerError(ctx, err)
		return
	}
	ctx.JSON(http.StatusNoContent, nil)

}

// CreateUserRequest represents the request format for creating a new user.
type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	NickName  string `json:"nick_name"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	Country   string `json:"country"`
}

// handlerError handles the error response for the http handlers
func handlerError(ctx *gin.Context, err error) {
	svcErr := new(errors.Error)
	if errors.As(err, &svcErr) {
		ctx.JSON(svcErr.Code, svcErr)
		return
	}
	ctx.JSON(http.StatusInternalServerError, errors.NewInternal(
		err.Error(),
	))
}

// Validate validates the fields of a CreateUserRequest. It returns an error if any field is invalid.
// we could have used a library like go-playground/validator for this purpose
func (c CreateUserRequest) Validate() error {
	details := []errors.Detail{}
	if c.FirstName == "" {
		details = append(details, errors.Detail{
			Field:       "first_name",
			Description: "first_name is required",
		})
	}

	if c.LastName == "" {
		details = append(details, errors.Detail{
			Field:       "last_name",
			Description: "last_name is required",
		})
	}

	if c.NickName == "" {
		details = append(details, errors.Detail{
			Field:       "nick_name",
			Description: "nick_name is required",
		})
	}

	if c.Password == "" {
		details = append(details, errors.Detail{
			Field:       "password",
			Description: "password is required",
		})
	}

	if c.Password != "" && (len(c.Password) < 8 || len(c.Password) > 40) {
		details = append(details, errors.Detail{
			Field:       "password",
			Description: "password must between 8 and 40 characters long",
		})
	}

	if c.Email == "" {
		details = append(details, errors.Detail{
			Field:       "email",
			Description: "email is required",
		})
	}

	if c.Email != "" && !emailRegex.MatchString(c.Email) {
		details = append(details, errors.Detail{
			Field:       "email",
			Description: "invalid email format",
		})
	}

	if len(c.Country) != 2 {
		details = append(details, errors.Detail{
			Field:       "country",
			Description: "country must be a 2-letter ISO country code",
		})
	}

	if len(details) > 0 {
		return errors.NewWrongInput("invalid user creation request", details...)
	}
	return nil
}

// UpdateUserRequest represents the request format for updating a user.
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	NickName  *string `json:"nick_name,omitempty"`
	Password  *string `json:"password,omitempty"`
	Email     *string `json:"email,omitempty"`
	Country   *string `json:"country,omitempty"`
}

func (c UpdateUserRequest) Vaildate() error {
	details := []errors.Detail{}
	if c.FirstName != nil && *c.FirstName == "" {
		details = append(details, errors.Detail{
			Field:       "first_name",
			Description: "first_name is required",
		})
	}

	if c.LastName != nil && *c.LastName == "" {
		details = append(details, errors.Detail{
			Field:       "last_name",
			Description: "last_name is required",
		})
	}

	if c.NickName != nil && *c.NickName == "" {
		details = append(details, errors.Detail{
			Field:       "nick_name",
			Description: "nick_name is required",
		})
	}

	if c.Password != nil && *c.Password == "" {
		details = append(details, errors.Detail{
			Field:       "password",
			Description: "password is required",
		})
	}

	if c.Password != nil && *c.Password != "" && (len(*c.Password) < 8 || len(*c.Password) > 40) {
		details = append(details, errors.Detail{
			Field:       "password",
			Description: "password must between 8 and 40 characters long",
		})
	}

	if c.Email != nil && *c.Email == "" {
		details = append(details, errors.Detail{
			Field:       "email",
			Description: "email is required",
		})
	}

	if c.Email != nil && *c.Email != "" && !emailRegex.MatchString(*c.Email) {
		details = append(details, errors.Detail{
			Field:       "email",
			Description: "invalid email format",
		})
	}

	if c.Country != nil && len(*c.Country) != 2 {
		details = append(details, errors.Detail{
			Field:       "country",
			Description: "country must be a 2-letter ISO country code",
		})
	}

	if len(details) > 0 {
		return errors.NewWrongInput("invalid user update request", details...)
	}
	return nil
}

// UserResponse represents the response format for a user. It doesn't include password, since is a sensitive information
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	NickName  string    `json:"nick_name"`
	Email     string    `json:"email"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsersResponse represents the response format for a list of users with its query parameters.
type ListUsersResponse struct {
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalRecords int64               `json:"total_records"`
	SortBy       string              `json:"sort_by"`
	SortOrder    string              `json:"sort_order"`
	Filters      map[string][]string `json:"filters"`
	Users        []UserResponse      `json:"users"`
}
