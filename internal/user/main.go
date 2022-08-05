package user

import "github.com/wolframdeus/noitifications-service/internal/timezone"

// Id описывает идентификатор пользователя ВКонтакте.
type Id uint64

type User struct {
	// Идентификатор пользователя.
	Id Id
	// Часовой пояс пользователя.
	Timezone timezone.Timezone
}

// New возвращает ссылку на новый экземпляр пользователя.
func New(id Id, tz timezone.Timezone) *User {
	return &User{Id: id, Timezone: tz}
}
