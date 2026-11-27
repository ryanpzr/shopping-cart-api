package additemtocart

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

// AddItem implements BR-CART-002. Adding a product already in the cart sums
// the quantities, re-validates the stock and refreshes the price snapshot.
func (u *usecase) AddItem(ctx context.Context, buyerID int, req AddItemRequest) (CartResponse, error) {
	if req.ProductID == nil {
		return CartResponse{}, apperrors.NewBadRequest("product_id is required")
	}
	if req.Quantity == nil {
		return CartResponse{}, apperrors.NewBadRequest("quantity is required")
	}
	if err := domain.ValidateQuantity(*req.Quantity); err != nil {
		return CartResponse{}, toAppError(err)
	}

	if err := u.ensureActiveCart(ctx, buyerID); err != nil {
		return CartResponse{}, toAppError(err)
	}

	err := u.uow.Do(ctx, func(r TxRepositories) error {
		cart, err := r.Carts.LockByBuyer(ctx, buyerID, domain.StatusActive)
		if err != nil {
			return err
		}

		product, err := r.Products.FindByID(ctx, *req.ProductID)
		if err != nil {
			return err
		}
		if !product.IsActive() {
			return productdomain.ErrProductNotFound
		}
		if product.SellerID == buyerID {
			return apperrors.NewForbidden("you cannot buy your own product")
		}

		item, exists := cart.ItemByProduct(product.ID)
		quantity := *req.Quantity
		if exists {
			quantity += item.Quantity
		}
		if err := domain.CheckStock(product.ID, product.Title, quantity, product.Quantity); err != nil {
			return err
		}

		item.Quantity = quantity
		item.PriceSnapshot = product.SalePrice()
		if exists {
			return r.Carts.UpdateItem(ctx, item)
		}
		item.CartID = cart.ID
		item.ProductID = product.ID
		_, err = r.Carts.AddItem(ctx, item)
		return err
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

// ensureActiveCart creates the buyer's active cart when there is none.
func (u *usecase) ensureActiveCart(ctx context.Context, buyerID int) error {
	_, err := u.repo.FindByBuyer(ctx, buyerID, domain.StatusActive)
	if !errors.Is(err, domain.ErrCartNotFound) {
		return err
	}
	_, err = u.repo.Create(ctx, buyerID, domain.StatusActive)
	if errors.Is(err, domain.ErrActiveCartExists) {
		return nil
	}
	return err
}

func toAppError(err error) error {
	var (
		appErr  *apperrors.AppError
		invalid *domain.ValidationError
		stock   *domain.StockError
	)
	switch {
	case errors.As(err, &appErr):
		return appErr
	case errors.As(err, &invalid):
		return apperrors.NewBadRequest(invalid.Message)
	case errors.As(err, &stock):
		return apperrors.NewConflict(stock.Error())
	case errors.Is(err, productdomain.ErrProductNotFound):
		return apperrors.NewNotFound("product not found")
	case errors.Is(err, domain.ErrCartNotFound):
		// The cart was checked out between ensureActiveCart and the lock.
		return apperrors.NewConflict("the cart changed, try again")
	default:
		return apperrors.NewInternalServer(err.Error())
	}
}
