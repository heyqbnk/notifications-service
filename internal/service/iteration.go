package service

import (
	"github.com/wolframdeus/noitifications-service/internal/app"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"sort"
	"sync"
	"time"
)

const (
	// SendNotificationUsersLimit - максимальное количество пользователей,
	// которым за раз можно отправить уведомление.
	SendNotificationUsersLimit = 100
)

type tasksTimezoneMap map[*task.Task][]timezone.Range

// Вызывает итерацию работы сервиса, которая подразумевает получение списка
// пользователей для всех задач, а также передачу их в задачи.
func (s *Service) RunIteration() {
	// Получаем текущий список всех часовых задач.
	tzRanges, tasksTzMap := s.getTimezonesMeta()

	var cursor = user.Id(0)
	for {
		// Порционно получаем список пользователей, удовлетворяющих условию по
		// часовым поясам.
		getResult, err := s.safeGetUsersByTimezones(tzRanges, cursor)
		if err != nil {
			// TODO: Здесь необходимо ещё несколько раз попробовать получить
			//  данные. Может быть соединение с провайдером моргнуло.
			return
		}

		// Пользователей нет, работать нам не с кем. Осуществляем ранний выход.
		if len(getResult.Users) == 0 {
			break
		}

		// Сортируем пользователей по возрастанию их часового пояса.
		sort.SliceStable(getResult.Users, func(i, j int) bool {
			return getResult.Users[i].Timezone < getResult.Users[j].Timezone
		})

		// Сортируем пользователей по задачам исходя из того, в каких часовых
		// поясах задача выполняется, а также исходя часового пояса пользователя.
		taskUsersMap := make(map[task.Id][]user.User, len(s.tasks))

		for t, timezones := range tasksTzMap {
			// Нам необходимо понять относительно которого часового пояса нужно
			// проводить сравнение для раннего выхода.
			comparedTimezone := timezones[len(timezones)-1].To

			// Пробегаемся по всем пользователям и проверяем, попадает ли их часовой
			// пояс под часовые пояса задачи.
			for _, u := range getResult.Users {
				// Если получилось так, что часовой пояс пользователя больше
				// максимального часового пояса задачи, делаем ранний выход.
				if comparedTimezone < u.Timezone {
					break
				}
				for _, tz := range timezones {
					if tz.ContainsTimezone(u.Timezone) {
						taskUsersMap[t.Id] = append(taskUsersMap[t.Id], u)
					}
				}
			}
		}

		// Получаем идентификаторы приложений и их задач для того, чтобы
		// распараллелить их дальнейшую обработку.
		appTasksMap := s.getAppTasksMap()
		var wg sync.WaitGroup

		// Пробегаемся по каждому приложению и для него выделяем отдельную
		// горутину, в которой будет выполняться обработка всех его задач.
		for _, tasks := range appTasksMap {
			wg.Add(1)

			go func(tasks []task.Task) {
				defer wg.Done()

				// Пробегаемся по каждой задаче и передаем в неё список подходящих
				// пользователей.
				for _, t := range tasks {
					// Если эта задача не зарегистрирована в карте с задачами и
					// пользователями, то подходящих для этой задачи пользователей просто
					// нет. Мы можем перейти к следующей задаче.
					users, ok := taskUsersMap[t.Id]
					if !ok {
						continue
					}

					// Если пользователей в задаче нет, переходим ко следующей.
					if len(users) == 0 {
						continue
					}

					// Передаём в задачу пользователей для проверки на отправку
					// уведомления.
					params, processErr := s.safeProcess(&t, users)
					if processErr != nil {
						continue
					}

					// Ничего не делаем в случае, если нет подходящих пользователей для
					// отправки уведомления.
					if len(params) == 0 {
						continue
					}

					// Отправляем уведомления пользователям.
					sendResult, err := s.sendNotifications(params)
					if err != nil {
						continue
					}

					// Сохраняем факт отправки уведомления.
					err = s.safeSaveSendResult(sendResult, t.AppId, t.Id, time.Now())
					if err != nil {
						continue
					}
				}
			}(tasks)
		}

		// Ожидаем выполнения всех горутин.
		wg.Wait()

		if !getResult.HasMore {
			break
		}
		// Перезаписываем курсор для следующего запроса.
		cursor = getResult.Cursor
	}
}

// Возвращает список интервалов часовых поясов, в которых должны находиться
// пользователи, чтобы попасть хотя бы в одну задачу.
func (s *Service) getTimezonesMeta() ([]timezone.Range, tasksTimezoneMap) {
	if len(s.tasks) == 0 {
		return []timezone.Range{}, tasksTimezoneMap{}
	}
	// Для начала создаем список интервалов часов поясов, пользователей в
	// которых нам необходимо получить.
	ranges := make([]timezone.Range, 0, len(s.tasks))
	tasksTzMap := make(tasksTimezoneMap, len(s.tasks))

	for _, t := range s.tasks {
		tTemp := t
		tasksTzMap[&tTemp] = t.GetTimezones()
		ranges = append(ranges, tasksTzMap[&tTemp]...)
	}

	// Сортируем массив интервалов по возрастанию их начала.
	sort.SliceStable(ranges, func(i, j int) bool {
		return ranges[i].From < ranges[j].From
	})

	// Пробегаемся по всем интервалам и склеиваем в случае необходимости.
	currentRange := ranges[0]
	minRanges := make([]timezone.Range, 0, len(ranges))

	for i := 1; i < len(ranges); i++ {
		r := ranges[i]

		if r.From <= currentRange.To {
			currentRange = *timezone.NewRange(currentRange.From, r.To)
			continue
		}
		minRanges = append(minRanges, currentRange)
		currentRange = r
	}
	minRanges = append(minRanges, currentRange)

	return minRanges, tasksTzMap
}

// Возвращает карту с ключом в виде идентификатора приложения и значением в
// виде списка задач, которые этому приложению принадлежат.
func (s *Service) getAppTasksMap() map[app.Id][]task.Task {
	res := make(map[app.Id][]task.Task)

	for _, t := range s.tasks {
		res[t.AppId] = append(res[t.AppId], t)
	}
	return res
}
