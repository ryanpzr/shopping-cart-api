package login

import "context"

type Usecase interface {
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
}
