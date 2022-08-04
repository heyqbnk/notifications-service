package mongodb

import (
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"time"
)

// UserId описывает идентификатор пользователя ВКонтакте.
type UserId uint64

type TaskId uint64

type Task struct {
	// Количество отправок этого уведомления пользователю.
	SendCount uint `bson:"sendCount"`
	// История отправки этого уведомления.
	History []time.Time `bson:"history"`
}

// Tasks описывает карту с информацией о каких-либо уведомлениях пользователя.
type Tasks map[TaskId]Task

type AppId uint64

type App struct {
	// Разрешена ли пользователю отправка уведомлений в этом приложении.
	AreNotificationsEnabled bool `bson:"areNotificationsEnabled"`
	// Информация об уведомлениях от этого приложения.
	Tasks Tasks `bson:"tasks"`
}

// Apps описывает карту с информацией о каких-либо приложениях пользователя.
type Apps map[AppId]App

type User struct {
	// Уникальный идентификатор пользователя ВКонтакте.
	Id UserId `bson:"_id"`
	// Информация о приложениях пользователя.
	Apps Apps `bson:"apps"`
	// Часовой пояс пользователя. Выражается в количестве минут, которое
	// необходимо прибавить ко времени по Гринвичу, чтобы получить локальное
	// время.
	Timezone int `bson:"timezone"`
}

// ToCommon конвертирует текущего пользователя к общему виду.
func (u *User) ToCommon() *user.User {
	return &user.User{
		Id:       user.Id(u.Id),
		Timezone: timezone.Timezone(u.Timezone),
	}
}

// NewUser создает ссылку на новый экземпляр User.
func NewUser(id UserId, apps Apps, timezone int) *User {
	return &User{Id: id, Apps: apps, Timezone: timezone}
}
