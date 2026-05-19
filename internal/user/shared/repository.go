package shared

import (
	"database/sql"

	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Repository interface {
	FindByEmail(email string) (*domain.User, error)
	Create(user domain.User) (domain.User, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(queryFindByEmail, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash,
		&u.Role, &u.Status, &u.TimeoutUntil,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewNotFound("user not found")
	}
	if err != nil {
		return nil, apperrors.NewInternalServer(err.Error())
	}
	return &u, nil
}

func (r *repository) Create(user domain.User) (domain.User, error) {
	err := r.db.QueryRow(queryCreate, user.Name, user.Email, user.PasswordHash).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.Role, &user.Status,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, apperrors.NewInternalServer(err.Error())
	}
	return user, nil
}
