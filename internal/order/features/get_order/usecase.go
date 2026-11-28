package getorder

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/order/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

const roleAdmin = "admin"

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

// Get implements BR-ORDER-002: a client sees only their own orders (403
// otherwise); an admin sees any order.
func (u *usecase) Get(ctx context.Context, orderID, userID int, role string) (OrderResponse, error) {
	order, err := u.repo.FindByID(ctx, orderID)
	if errors.Is(err, domain.ErrOrderNotFound) {
		return OrderResponse{}, apperrors.NewNotFound("order not found")
	}
	if err != nil {
		return OrderResponse{}, apperrors.NewInternalServer(err.Error())
	}

	if role != roleAdmin && !order.OwnedBy(userID) {
		return OrderResponse{}, apperrors.NewForbidden("you are not the owner of this order")
	}
	return toResponse(order), nil
}
