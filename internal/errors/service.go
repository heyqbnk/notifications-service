package errors

type ServiceError struct {
	// Оригинальная выброшенная ошибка.
	Original error
}

// NewServiceError возвращает ссылку на новый экземпляр ServiceError.
func NewServiceError(err error) *ServiceError {
	return &ServiceError{
		Original: err,
	}
}
