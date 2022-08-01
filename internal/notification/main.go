package notification

import "github.com/wolframdeus/noitifications-service/internal/user"

type Params struct {
	// Идентификатор пользователя которому надо отправить уведомление.
	UserId user.Id
	// Текст уведомления.
	Message string
}
