package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/zechao/faceit-user-svc/event"
	"github.com/zechao/faceit-user-svc/log"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/user"
)

type userService struct {
	userRepo     user.Repository
	eventHandler event.EventHandler
}

// NewUserService creates a new user service with the provided user repository and event handler.
func NewUserService(userRepo user.Repository, eventHandler event.EventHandler) user.Service {
	return &userService{
		userRepo:     userRepo,
		eventHandler: eventHandler,
	}
}

// CreateUser implements user.Service. It will generate a new UUID for the user and save it to the repository.
func (ur *userService) CreateUser(ctx context.Context, input *user.CreateUserInput) (*user.User, error) {
	u, err := user.NewUser(input)
	if err != nil {
		return nil, fmt.Errorf("fail creating user %w", err)
	}

	log.Info(ctx, "creating new user", slog.String(
		"user_id", u.ID.String(),
	))

	res, err := ur.userRepo.CreateUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("fail creating user %w", err)
	}

	err = ur.eventHandler.SendEvent(ctx, string(user.UserCreated), res.ID)
	if err != nil {
		return nil, fmt.Errorf("fail sending event %w", err)
	}

	return res, nil
}

// UpdateUser implements will check if the user exists before updating.
// return not found error if not exist and return duplicated error if the update cause a duplicated key.
// otherwise update the user and return the updated user.
func (ur *userService) UpdateUser(ctx context.Context, input *user.UpdateUserInput) (*user.User, error) {
	log.Info(ctx, "updating user", slog.String(
		"user_id", input.ID.String(),
	))
	
	userToUpdate, err := ur.userRepo.GetUserByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	err = userToUpdate.Update(input)
	if err != nil {
		log.Error(ctx, "failed to update user", slog.String(
			"user_id", userToUpdate.ID.String(),
		))
		return nil, err
	}

	res, err := ur.userRepo.UpdateUser(ctx, userToUpdate)
	if err != nil {
		return nil, err
	}

	err = ur.eventHandler.SendEvent(ctx, string(user.UserUpdated), res.ID)
	if err != nil {
		return nil, fmt.Errorf("fail sending event %w", err)
	}
	return res, nil
}

// DeleteUser will delete the user from the repository by ID.
// If the user does not exist do nothing
func (ur *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	log.Info(ctx, "deleting user", slog.String(
		"user_id", id.String(),
	))
	err := ur.userRepo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	err = ur.eventHandler.SendEvent(ctx, string(user.UserDeleted), id)
	if err != nil {
		return fmt.Errorf("fail sending event %w", err)
	}
	return nil
}

// ListUsers lists the users from the repository based on the query.
// If no users are found, it will return an empty slice and nil error
// If the page is out of range, it will return an empty slice and no error.
func (ur *userService) ListUsers(ctx context.Context, q query.Query) (*query.PaginationResponse[user.User], error) {
	log.Info(ctx, "listing user with query", slog.Any(
		"query", q,
	))
	count, err := ur.userRepo.CountUsers(ctx, q.Filters)
	if err != nil {
		return nil, err
	}

	totalPages := count / int64(q.PageSize)
	if count%int64(q.PageSize) > 0 {
		totalPages++
	}

	res := query.PaginationResponse[user.User]{
		Page:         q.Page,
		PageSize:     q.PageSize,
		SortBy:       q.SortBy,
		SortOrder:    q.SortOrder,
		Filters:      q.Filters,
		TotalRecords: count,
		Data:         []user.User{},
	}

	if count == 0 || int64(q.Page) > totalPages {
		return &res, nil
	}
	users, err := ur.userRepo.ListUsers(ctx, q)
	if err != nil {
		return nil, err
	}
	res.Data = users
	return &res, nil
}

var _ user.Service = (*userService)(nil)
