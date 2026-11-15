package togglestatus_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryanpzr/shopping-cart-api/internal/product/domain"
	togglestatus "github.com/ryanpzr/shopping-cart-api/internal/product/features/toggle_status"
	"github.com/ryanpzr/shopping-cart-api/internal/product/shared/mocks"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

func assertAppError(t *testing.T, err error, code int) {
	t.Helper()
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
}

func TestToggle(t *testing.T) {
	tests := []struct {
		from, to string
	}{
		{domain.StatusActive, domain.StatusInactive},
		{domain.StatusInactive, domain.StatusActive},
	}
	for _, tt := range tests {
		t.Run(tt.from+" to "+tt.to, func(t *testing.T) {
			repo := new(mocks.Repository)
			repo.On("FindByID", mock.Anything, 3).Return(&domain.Product{ID: 3, SellerID: 7, Status: tt.from}, nil)
			repo.On("UpdateStatus", mock.Anything, 3, tt.to).Return(nil)

			resp, err := togglestatus.NewUsecase(repo).Toggle(context.Background(), 3, 7)

			require.NoError(t, err)
			assert.Equal(t, togglestatus.ToggleResponse{ID: 3, Status: tt.to}, resp)
			repo.AssertExpectations(t)
		})
	}
}

func TestToggle_NotOwner(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(&domain.Product{ID: 3, SellerID: 7, Status: domain.StatusActive}, nil)

	_, err := togglestatus.NewUsecase(repo).Toggle(context.Background(), 3, 8)

	assertAppError(t, err, http.StatusForbidden)
	repo.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
}

func TestToggle_NotFound(t *testing.T) {
	repo := new(mocks.Repository)
	repo.On("FindByID", mock.Anything, 3).Return(nil, domain.ErrProductNotFound)

	_, err := togglestatus.NewUsecase(repo).Toggle(context.Background(), 3, 7)

	assertAppError(t, err, http.StatusNotFound)
}
