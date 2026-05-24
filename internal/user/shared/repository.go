package shared

import (
	"database/sql"
	"time"

	"github.com/ryanpzr/shopping-cart-api/internal/user/domain"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
)

type Repository interface {
	FindByEmail(email string) (*domain.User, error)
	Create(user domain.User) (domain.User, error)
	FindByID(id int) (*domain.User, error)
	FindAll(limit, offset int) ([]domain.User, int, error)
	UpdateProfile(id int, name, email string) (domain.User, error)
	UpdateStatus(id int, status string, timeoutUntil *time.Time) (domain.User, error)
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

func (r *repository) FindByID(id int) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(queryFindByID, id).Scan(
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

func (r *repository) FindAll(limit, offset int) ([]domain.User, int, error) {
	var total int
	if err := r.db.QueryRow(queryCountAll).Scan(&total); err != nil {
		return nil, 0, apperrors.NewInternalServer(err.Error())
	}

	rows, err := r.db.Query(queryFindAll, limit, offset)
	if err != nil {
		return nil, 0, apperrors.NewInternalServer(err.Error())
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email, &u.PasswordHash,
			&u.Role, &u.Status, &u.TimeoutUntil,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, apperrors.NewInternalServer(err.Error())
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.NewInternalServer(err.Error())
	}

	return users, total, nil
}

func (r *repository) UpdateProfile(id int, name, email string) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(queryUpdateProfile, name, email, id).Scan(
		&u.ID, &u.Name, &u.Email,
		&u.Role, &u.Status,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.User{}, apperrors.NewNotFound("user not found")
	}
	if err != nil {
		return domain.User{}, apperrors.NewInternalServer(err.Error())
	}
	return u, nil
}

func (r *repository) UpdateStatus(id int, status string, timeoutUntil *time.Time) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(queryUpdateStatus, status, timeoutUntil, id).Scan(
		&u.ID, &u.Name, &u.Email,
		&u.Role, &u.Status, &u.TimeoutUntil,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.User{}, apperrors.NewNotFound("user not found")
	}
	if err != nil {
		return domain.User{}, apperrors.NewInternalServer(err.Error())
	}
	return u, nil
}
