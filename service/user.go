package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zechao/faceit-user-svc/query"
	"github.com/zechao/faceit-user-svc/user"
)

type userService struct {
	userRepo user.Repository
}

func NewUserService(userRepo user.Repository) user.Service {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser implements user.Service. It will generate a new UUID for the user and save it to the repository.
func (ur *userService) CreateUser(ctx context.Context, u *user.User) (*user.User, error) {
	u.ID = uuid.New()
	err := ur.userRepo.CreateUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUser implements will check if the user exists before updating.
// return not found error if not exist, otherwise update the user and return the updated user.
func (ur *userService) UpdateUser(ctx context.Context, u *user.User) (*user.User, error) {
	panic("unimplemented")
}

// DeleteUser implements user.Service. It will delete the user from the repository by ID. If the user does not exist do nothing
func (ur *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return ur.userRepo.DeleteUser(ctx, id)

}

// ListUsers implements user.Service.
func (ur *userService) ListUsers(ctx context.Context, query query.Query) (*query.PaginationResponse[user.User], error) {
	
}

var _ user.Service = (*userService)(nil)
