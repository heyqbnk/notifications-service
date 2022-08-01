package service

import (
	"fmt"
	"github.com/wolframdeus/noitifications-service/internal/app"
	"github.com/wolframdeus/noitifications-service/internal/user"
)

// В безопасном режиме изменяет разрешение на отправку уведомлений пользователю.
func (s *Service) safeSetAllowStatusForUser(
	userId user.Id,
	appId app.Id,
	allowed bool,
) (err error) {
	defer func() {
		if e := recover(); e != nil {
			var text string
			if eConverted, ok := e.(error); ok {
				text = eConverted.Error()
			} else {
				text = fmt.Sprintf("%s", e)
			}
			err = fmt.Errorf("произошла ошибка во время изменения разрешения на отправку уведомления: %s", text)
		}
		// Логируем ошибку.
		if err != nil {
			_ = s.safeOnError(err)
		}
	}()

	err = s.provider.SetAllowStatusForUser(userId, appId, allowed)
	return err
}

// В безопасном режиме вызывает функцию логирования ошибки.
func (s *Service) safeOnError(err error) (e error) {
	if s.onError == nil {
		return
	}
	defer func() {
		if eRecovered := recover(); eRecovered != nil {
			var text string
			if eConverted, ok := eRecovered.(error); ok {
				text = eConverted.Error()
			} else {
				text = fmt.Sprintf("%s", eRecovered)
			}
			err = fmt.Errorf("произошла ошибка во время логирования внутренней ошибки: %s", text)
		}
	}()

	s.onError(err)
	return nil
}
