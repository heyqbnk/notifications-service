package service

import (
	"github.com/wolframdeus/noitifications-service/internal/notification"
	"github.com/wolframdeus/noitifications-service/internal/user"
)

// Выполняет отправку уведомлений пользователям.
// TODO: Возвращать не идентификаторы пользователей, а структуры с описанием
//  результата выполнения для каждого пользователя.
func (s *Service) sendNotifications(params []notification.Params) ([]user.Id, error) {
	// "Привет Вася!", [[1, 2, 3], [92, 11, 2983, 22]]
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

	successUserIds := make([]user.Id, 0, len(params))

	// Пробегаемся по каждой пачке и рассылаем уведомления.
	for message, userIds := range batches {
		for _, b := range userIds {
			// TODO: Скорее всего это можно делать в отдельных горутинах.
			res, _ := s.vk.NotificationsSendMessage(map[string]interface{}{
				"user_ids": b,
				"message":  message,
			})

			// TODO: Error handling.
			for _, r := range res {
				if r.Status {
					successUserIds = append(successUserIds, user.Id(r.UserID))
				}
			}
		}
	}
	return successUserIds, nil
}
