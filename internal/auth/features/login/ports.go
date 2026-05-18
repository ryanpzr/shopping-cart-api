package login

type Usecase interface {
	Login(req LoginRequest) (LoginResponse, error)
}
