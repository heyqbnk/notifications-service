package providers

import (
	"github.com/wolframdeus/noitifications-service/internal/app"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"time"
)

type GetUsersByTimezonesResult struct {
	// Текущее положение курсора в выдаче.
	Cursor user.Id
	// Список пользователей, удовлетворяющих условию.
	Users []user.User
	// Флаг говорящий о том, что пользователей удовлетворяющих условию больше
	// чем драйвер смог вернуть.
	HasMore bool
}

// NewGetUsersByTimezonesResult создает ссылку на новый
// экземпляр GetUsersByTimezonesResult.
func NewGetUsersByTimezonesResult(cursor user.Id, users []user.User, hasMore bool) *GetUsersByTimezonesResult {
	return &GetUsersByTimezonesResult{Cursor: cursor, Users: users, HasMore: hasMore}
}

type Provider interface {
	// GetUsersByTimezones возвращает пользователей удовлетворяющих условию
	// наличия часового пояса.
	GetUsersByTimezones(tz []timezone.Range, cursor user.Id) (*GetUsersByTimezonesResult, error)
	// SetAllowStatusForUser - функция для изменения разрешения на отправку
	// уведомлений пользователю.
	SetAllowStatusForUser(userId user.Id, appId app.Id, allowed bool) error
	// SaveNotificationDate сохраняет дату отправки указанного уведомления.
	SaveNotificationDate(
		userIds []user.Id,
		appId app.Id,
		taskId task.Id,
		date time.Time,
	) error
	// UserExists возвращает true в случае, если пользователь зарегистрирован в
	// сервисе.
	UserExists(userId user.Id) (bool, error)
	// RegisterUser регистрирует пользователя в сервисе.
	RegisterUser(userId user.Id, tz timezone.Timezone) error
}
