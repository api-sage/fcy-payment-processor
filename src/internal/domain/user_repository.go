package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, user User) (User, error)
}
