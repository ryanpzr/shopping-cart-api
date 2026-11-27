package cart

import (
	"github.com/gin-gonic/gin"
	additemtocart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/add_item_to_cart"
	changeitemquantity "github.com/ryanpzr/shopping-cart-api/internal/cart/features/change_item_quantity"
	"github.com/ryanpzr/shopping-cart-api/internal/cart/features/checkout"
	getcart "github.com/ryanpzr/shopping-cart-api/internal/cart/features/get_cart"
	removeitem "github.com/ryanpzr/shopping-cart-api/internal/cart/features/remove_item"
)

type CartHandlers struct {
	Get            getcart.Handler
	AddItem        additemtocart.Handler
	ChangeQuantity changeitemquantity.Handler
	RemoveItem     removeitem.Handler
	Checkout       checkout.Handler
}

// MapRouters registers the cart routes under /carts/me. The caller must apply
// middleware.Auth(): the cart always belongs to the authenticated user.
func MapRouters(r *gin.RouterGroup, h CartHandlers) {
	r.GET("", h.Get.Get)
	r.POST("/items", h.AddItem.AddItem)
	r.PATCH("/items/:itemId", h.ChangeQuantity.ChangeQuantity)
	r.DELETE("/items/:itemId", h.RemoveItem.RemoveItem)
	r.POST("/checkout", h.Checkout.Checkout)
}
