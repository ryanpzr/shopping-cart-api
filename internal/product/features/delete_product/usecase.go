package deleteproduct

import "github.com/ryanpzr/shopping-cart-api/pkg/apperrors"

type usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Delete(productID, userID int, role string) error {
	p, err := u.repo.FindById(productID)
	if err != nil {
		return err
	}

	// Verificar permissão: admin pode deletar qualquer produto; client só o próprio
	if role != "admin" && p.SellerID != userID {
		return apperrors.NewForbidden("you are not the owner of this product")
	}

	// TODO(module-05): HasActiveOrders depende da tabela 'orders' criada em sdd/specs/modules/05-order.md.
	// Quando o módulo 05 for implementado, esta verificação passará a funcionar em runtime.
	hasOrders, err := u.repo.HasActiveOrders(productID)
	if err != nil {
		return err
	}
	if hasOrders {
		return apperrors.NewConflict("product has active orders and cannot be deleted")
	}

	if err := u.repo.Delete(productID); err != nil {
		return err
	}

	// TODO(module-06): Emitir evento 'product_deleted' no activity log.
	// Ver: sdd/specs/modules/06-activity-log.md

	return nil
}
