package register

type Usecase interface {
	Register(req RegisterRequest) (RegisterResponse, error)
}
