package providers

import (
	"github.com/wolframdeus/noitifications-service/internal/app"
	customerror "github.com/wolframdeus/noitifications-service/internal/errors"
	"github.com/wolframdeus/noitifications-service/internal/notification"
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
func NewGetUsersByTimezonesResult(
	cursor user.Id,
	users []user.User,
	hasMore bool,
) *GetUsersByTimezonesResult {
	return &GetUsersByTimezonesResult{Cursor: cursor, Users: users, HasMore: hasMore}
}

type Provider interface {
	// GetUsersByTimezones возвращает пользователей удовлетворяющих условию
	// наличия часового пояса.
	GetUsersByTimezones(tz []timezone.Range, cursor user.Id) (*GetUsersByTimezonesResult, *customerror.ServiceError)

	// SetAllowStatusForUser - функция для изменения разрешения на отправку
	// уведомлений пользователю.
	SetAllowStatusForUser(userId user.Id, appId app.Id, allowed bool) *customerror.ServiceError

	// SaveSendResult сохраняет результаты отправки уведомлений.
	SaveSendResult(
		results *notification.SendResult,
		appId app.Id,
		taskId task.Id,
		date time.Time,
	) *customerror.ServiceError
}
