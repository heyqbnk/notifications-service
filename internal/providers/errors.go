package providers

import "errors"

var (
	ErrUserDoesNotExist  = errors.New("пользователь не существует")
	ErrUserAlreadyExists = errors.New("пользователь уже существует")
)
