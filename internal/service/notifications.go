package service

import (
	"github.com/wolframdeus/noitifications-service/internal/errors"
	"github.com/wolframdeus/noitifications-service/internal/notification"
	"github.com/wolframdeus/noitifications-service/internal/user"
)

// Выполняет отправку уведомлений пользователям.
func (s *Service) sendNotifications(
	params []notification.Params,
) (*notification.SendResult, *errors.ServiceError) {
	// Создаем карту, в которой в качестве ключа будет сообщение, а в качестве
	// значения - список батчей из идентификаторов пользователей.
	// Пример: { "Привет Вася!": [[1, 2, 3], [92, 11, 2983, 22]] }
	batches := make(map[string][][]user.Id)

	for _, p := range params {
		// Отрезаем все символы после 256-ого и вставляем в конце 3 точки. Это
		// единственное адекватное решение, которые мы здесь можем использовать.
		if len(p.Message) > 256 {
			p.Message = p.Message[0:253] + "..."
		}

		// Получаем список всех пользователей с таким сообщением.
		userIds, ok := batches[p.Message]
		if !ok {
			batches[p.Message] = [][]user.Id{{p.UserId}}
			continue
		}

		// Получаем последний пачку с мыслью о том, что туда можно будет добавить
		// этого пользователя.
		batch := userIds[len(userIds)-1]

		// Если эта пачка уже переполнена, то мы добавляем новую.
		if len(batch) == SendNotificationUsersLimit {
			batches[p.Message] = append(batches[p.Message], []user.Id{p.UserId})
			continue
		}
		userIds[len(userIds)-1] = append(batch, p.UserId)
	}

	var result *notification.SendResult

	// Пробегаемся по каждой пачке и рассылаем уведомления.
	for message, userIds := range batches {
		for _, b := range userIds {
			// TODO: Скорее всего это можно делать в отдельных горутинах.
			// TODO: Не используется fragment :(
			res, err := s.vk.NotificationsSendMessage(map[string]interface{}{
				"user_ids": b,
				"message":  message,
			})

			// Если произошла ошибка внутреннего характера, добавляем пользователей
			// в соответствующий раздел.
			if err != nil {
				result.InternalError = append(result.InternalError, b...)
			}

			// Пробегаемся по каждому пользователю и добавляем его в свой раздел.
			for _, r := range res {
				uid := user.Id(r.UserID)

				if r.Status {
					result.Success = append(result.Success, uid)
				} else {
					// Спецификация ошибок:
					// https://dev.vk.com/method/notifications.sendMessage#Результат
					switch r.Error.Code {
					case 1, 4:
						result.NotificationsDisabled = append(result.NotificationsDisabled, uid)
					case 2:
						result.HourRateLimitReached = append(result.HourRateLimitReached, uid)
					case 3:
						result.DayRateLimitReached = append(result.DayRateLimitReached, uid)
					default:
						result.UnknownError = append(result.UnknownError, uid)
					}
				}
			}
		}
	}

	// TODO: Как-то логировать ошибки, которые возвращаются от API ВКонтакте,
	//  чтобы понимать, что что-то не так.

	return result, nil
}
