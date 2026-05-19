package apperrors

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string { return e.Message }

func NewBadRequest(msg string) *AppError      { return &AppError{400, msg} }
func NewUnauthorized(msg string) *AppError    { return &AppError{401, msg} }
func NewForbidden(msg string) *AppError       { return &AppError{403, msg} }
func NewNotFound(msg string) *AppError        { return &AppError{404, msg} }
func NewConflict(msg string) *AppError        { return &AppError{409, msg} }
func NewTooManyRequests(msg string) *AppError { return &AppError{429, msg} }
func NewInternalServer(msg string) *AppError  { return &AppError{500, msg} }
