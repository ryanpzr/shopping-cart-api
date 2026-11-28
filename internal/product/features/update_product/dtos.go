package updateproduct

import (
	"encoding/json"
	"time"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
)

// Optional tells an omitted JSON field apart from an explicit null:
// omitted → Set is false; null → Set is true and Value is nil.
type Optional[T any] struct {
	Set   bool
	Value *T
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}

// UpdateRequest is a partial update: omitted fields keep their current value.
// photo, description and category can be cleared with null; the other fields
// cannot be null. seller_id and status are not accepted here.
type UpdateRequest struct {
	Photo              Optional[string]  `json:"photo"`
	Title              Optional[string]  `json:"title"`
	Description        Optional[string]  `json:"description"`
	Price              Optional[float64] `json:"price"`
	DiscountPercentage Optional[int]     `json:"discount_percentage"`
	Quantity           Optional[int]     `json:"quantity"`
	Category           Optional[string]  `json:"category"`
}

func (r UpdateRequest) isEmpty() bool {
	return len(r.changedFields()) == 0
}

// changedFields lists the JSON names of the fields present in the request, in
// alphabetical order. It is the changed_fields metadata of the audit event.
func (r UpdateRequest) changedFields() []string {
	fields := []string{}
	for _, f := range []struct {
		name string
		set  bool
	}{
		{"category", r.Category.Set},
		{"description", r.Description.Set},
		{"discount_percentage", r.DiscountPercentage.Set},
		{"photo", r.Photo.Set},
		{"price", r.Price.Set},
		{"quantity", r.Quantity.Set},
		{"title", r.Title.Set},
	} {
		if f.set {
			fields = append(fields, f.name)
		}
	}
	return fields
}

type ProductResponse struct {
	ID                 int       `json:"id"`
	SellerID           int       `json:"seller_id"`
	Photo              *string   `json:"photo"`
	Title              string    `json:"title"`
	Description        *string   `json:"description"`
	Price              float64   `json:"price"`
	DiscountPercentage int       `json:"discount_percentage"`
	FinalPrice         float64   `json:"final_price"`
	Quantity           int       `json:"quantity"`
	Status             string    `json:"status"`
	Category           *string   `json:"category"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func toResponse(p domain.Product) ProductResponse {
	return ProductResponse{
		ID:                 p.ID,
		SellerID:           p.SellerID,
		Photo:              p.Photo,
		Title:              p.Title,
		Description:        p.Description,
		Price:              p.Price,
		DiscountPercentage: p.DiscountPercentage,
		FinalPrice:         p.FinalPrice(),
		Quantity:           p.Quantity,
		Status:             p.Status,
		Category:           p.Category,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}
