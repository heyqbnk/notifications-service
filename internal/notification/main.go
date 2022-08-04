package notification

import "github.com/wolframdeus/noitifications-service/internal/user"

type Params struct {
	// Идентификатор пользователя которому надо отправить уведомление.
	UserId user.Id
	// Текст уведомления.
	Message string
	// Фрагмент, который необходимо использовать в уведомлении.
	Fragment string
}

type SendResult struct {
	// Список пользователей, отправка уведомлений которым была успешна.
	Success []user.Id
	// Список пользователей, отправка уведомлений которым запрещена.
	NotificationsDisabled []user.Id
	// Список пользователей, ошибка у которых не попадает ни под одну из
	// категорий.
	UnknownError []user.Id
	// Список пользователей, у которых достигнут лимит по уведомлениям за
	// последний час.
	HourRateLimitReached []user.Id
	// Список пользователей, у которых достигнут лимит по уведомлениям за
	// текущий день.
	DayRateLimitReached []user.Id
	// Список пользователей, которым не удалось отправить уведомление ввиду
	// внутренней ошибки.
	InternalError []user.Id
}
