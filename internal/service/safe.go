package service

import (
	"errors"
	"fmt"
	"github.com/wolframdeus/noitifications-service/internal/app"
	customerror "github.com/wolframdeus/noitifications-service/internal/errors"
	"github.com/wolframdeus/noitifications-service/internal/notification"
	"github.com/wolframdeus/noitifications-service/internal/providers"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"time"
)

// Создает ссылку на экземпляр ServiceError из восстановленной ошибки.
func (s *Service) recoverServiceError(e interface{}) *customerror.ServiceError {
	var eRecovered error

	if eConverted, ok := e.(error); ok {
		eRecovered = eConverted
	} else {
		eRecovered = errors.New(fmt.Sprintf("%s", e))
	}
	return customerror.NewServiceError(eRecovered)
}

// В безопасном режиме вызывает функцию SetAllowStatusForUser провайдера.
func (s *Service) safeSetAllowStatusForUser(
	userId user.Id,
	appId app.Id,
	allowed bool,
) (err *customerror.ServiceError) {
	defer func() {
		if e := recover(); e != nil {
			err = s.recoverServiceError(e)
		}

		// Если ошибка произошла, захватываем её и наполняем контекстными данными.
		if err != nil {
			s.captureServiceError(err, &CaptureOptions{
				Contexts: map[string]interface{}{
					"Parameters": map[string]interface{}{
						"userId":  userId,
						"appId":   appId,
						"allowed": allowed,
					},
				},
			})
		}
	}()

	err = s.provider.SetAllowStatusForUser(userId, appId, allowed)
	return
}

// В безопасном режиме вызывает функцию GetUsersByTimezones провайдера.
func (s *Service) safeGetUsersByTimezones(
	tz []timezone.Range,
	cursor user.Id,
) (res *providers.GetUsersByTimezonesResult, err *customerror.ServiceError) {
	defer func() {
		if e := recover(); e != nil {
			err = s.recoverServiceError(e)
		}

		// Если ошибка произошла, захватываем её и наполняем контекстными данными.
		if err != nil {
			s.captureServiceError(err, &CaptureOptions{
				Contexts: map[string]interface{}{
					"Parameters": map[string]interface{}{
						"tz":     tz,
						"cursor": cursor,
					},
				},
			})
		}
	}()

	res, err = s.provider.GetUsersByTimezones(tz, cursor)
	return
}

// В безопасном режиме вызывает функцию SaveSendResult провайдера.
func (s *Service) safeSaveSendResult(
	results *notification.SendResult,
	appId app.Id,
	taskId task.Id,
	date time.Time,
) (err *customerror.ServiceError) {
	defer func() {
		if e := recover(); e != nil {
			err = s.recoverServiceError(e)
		}

		// Если ошибка произошла, захватываем её и наполняем контекстными данными.
		if err != nil {
			s.captureServiceError(err, &CaptureOptions{
				Contexts: map[string]interface{}{
					"Parameters": map[string]interface{}{
						"results": results,
						"appId":   appId,
						"taskId":  taskId,
						"date":    date,
					},
				},
			})
		}
	}()

	err = s.provider.SaveSendResult(results, appId, taskId, date)
	return
}

// В безопасном режиме вызывает функцию Process задачи.
func (s *Service) safeProcess(
	t *task.Task,
	users []user.User,
) (res []notification.Params, err *customerror.TaskError) {
	defer func() {
		if e := recover(); e != nil {
			var eRecovered error

			if eConverted, ok := e.(error); ok {
				eRecovered = eConverted
			} else {
				eRecovered = errors.New(fmt.Sprintf("%s", e))
			}
			err = customerror.NewTaskError(t.AppId, t.Id, eRecovered)
		}

		// Если ошибка произошла, захватываем её и наполняем контекстными данными.
		if err != nil {
			s.captureTaskError(err, &CaptureOptions{
				Contexts: map[string]interface{}{
					"Parameters": map[string]interface{}{
						"users": users,
					},
				},
			})
		}
	}()

	res, err = t.Process(users)
	return
}
