package changeitemquantity

import (
	"context"
	"errors"

	"github.com/ryanpzr/shopping-cart-api/internal/cart/domain"
	productdomain "github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type usecase struct {
	repo Repository
	uow  UnitOfWork
}

func NewUsecase(repo Repository, uow UnitOfWork) Usecase {
	return &usecase{repo: repo, uow: uow}
}

// ChangeQuantity implements BR-CART-004: sets the item quantity after checking
// the current stock. The price snapshot is kept.
func (u *usecase) ChangeQuantity(ctx context.Context, buyerID, itemID int, req ChangeQuantityRequest) (CartResponse, error) {
	if req.Quantity == nil {
		return CartResponse{}, apperrors.NewBadRequest("quantity is required")
	}
	if err := domain.ValidateQuantity(*req.Quantity); err != nil {
		return CartResponse{}, toAppError(err)
	}

	err := u.uow.Do(ctx, func(r TxRepositories) error {
		cart, err := r.Carts.LockByBuyer(ctx, buyerID, domain.StatusActive)
		if errors.Is(err, domain.ErrCartNotFound) {
			return domain.ErrItemNotFound
		}
		if err != nil {
			return err
		}

		item, ok := cart.ItemByID(itemID)
		if !ok {
			return domain.ErrItemNotFound
		}

		product, err := r.Products.FindByID(ctx, item.ProductID)
		if err != nil {
			return err
		}
		if err := domain.CheckStock(product.ID, product.Title, *req.Quantity, product.Quantity); err != nil {
			return err
		}

		item.Quantity = *req.Quantity
		return r.Carts.UpdateItem(ctx, item)
	})
	if err != nil {
		return CartResponse{}, toAppError(err)
	}

	cart, err := u.repo.FindByBuyer(ctx, buyerID, domain.StatusActive)
	if err != nil {
		return CartResponse{}, toAppError(err)
	}
	return toResponse(cart), nil
}

func toAppError(err error) error {
	var (
		invalid *domain.ValidationError
		stock   *domain.StockError
	)
	switch {
	case errors.As(err, &invalid):
		return apperrors.NewBadRequest(invalid.Message)
	case errors.As(err, &stock):
		return apperrors.NewConflict(stock.Error())
	case errors.Is(err, domain.ErrItemNotFound), errors.Is(err, domain.ErrCartNotFound):
		return apperrors.NewNotFound("cart item not found")
	case errors.Is(err, productdomain.ErrProductNotFound):
		return apperrors.NewNotFound("product not found")
	default:
		return apperrors.NewInternalServer(err.Error())
	}
}
