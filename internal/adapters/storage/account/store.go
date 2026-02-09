package account

import (
	"context"

	domain "workshop/internal/domain/account"
)

// Store persists Account state.
type Store interface {
	GetByID(ctx context.Context, id string) (domain.Account, error)
	GetByEmail(ctx context.Context, email string) (domain.Account, error)
	Save(ctx context.Context, value domain.Account) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Account, error)
	Count(ctx context.Context) (int, error)
	SaveActivationToken(ctx context.Context, token domain.ActivationToken) error
	GetActivationTokenByToken(ctx context.Context, token string) (domain.ActivationToken, error)
	InvalidateTokensForAccount(ctx context.Context, accountID string) error
}

// ListFilter carries filtering parameters for List operations.
type ListFilter struct {
	Limit  int
	Offset int
	Role   string
}
