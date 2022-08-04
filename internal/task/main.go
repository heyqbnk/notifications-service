package task

import (
	"errors"
	"fmt"
	"github.com/wolframdeus/noitifications-service/internal"
	"github.com/wolframdeus/noitifications-service/internal/app"
	customerror "github.com/wolframdeus/noitifications-service/internal/errors"
	"github.com/wolframdeus/noitifications-service/internal/notification"
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"sort"
)

const (
	// Количество минут в одном дне.
	dayMinutes = 24 * 60
)

type Id uint

type ProcessFunc func(users []user.User) ([]notification.Params, *customerror.TaskError)

// Task описывает структуру любой задачи-уведомления.
type Task struct {
	// Идентификатор приложения-владельца.
	AppId app.Id
	// Идентификатор самой задачи.
	Id Id
	// Начало временного промежутка для отправки этого уведомления. Данное
	// значение описывает локальное время пользователя.
	From *internal.Time
	// Конец временного промежутка для отправки этого уведомления. Данное
	// значение описывает локальное время пользователя.
	To      *internal.Time
	process ProcessFunc
}

// GetTimezones возвращает массив диапазонов часовых поясов, которые
// соответствуют указанному в задаче времени. Результирующий массив
// отсортирован по возрастанию концов интервалов.
func (s *Task) GetTimezones() []timezone.Range {
	fromTz := s.From.GetTimezones()
	toTz := s.To.GetTimezones()

	if len(fromTz) != len(toTz) {
		// Дополняем массивы часовых поясов необходимыми для создания интервалов
		// значениями.
		if len(fromTz) != 2 {
			fromTz = append(fromTz, timezone.CutTimezone(fromTz[0]-dayMinutes))
			fromTz[1], fromTz[0] = fromTz[0], fromTz[1]
		}
		if len(toTz) != 2 {
			toTz = append(toTz, timezone.CutTimezone(toTz[0]+dayMinutes))
		}
	}
	res := make([]timezone.Range, len(fromTz))

	for i := 0; i < len(fromTz); i++ {
		res[i] = *timezone.NewRange(fromTz[i], toTz[i])
	}

	// Если получилось так, что количество интервалов стало больше 1, мы
	// отсортируем их по возрастанию их концов.
	if len(res) > 1 {
		sort.SliceStable(res, func(i, j int) bool {
			return res[i].To < res[j].To
		})
	}
	return res
}

// Process принимает на вход список пользователей и проверяет, необходимо ли
// им и с какими параметрами отправить уведомление.
func (s *Task) Process(users []user.User) (params []notification.Params, err *customerror.TaskError) {
	defer func() {
		if e := recover(); e != nil {
			var eRecovered error

			if eConverted, ok := e.(error); ok {
				eRecovered = eConverted
			} else {
				eRecovered = errors.New(fmt.Sprintf("%s", e))
			}
			err = customerror.NewTaskError(s.AppId, s.Id, eRecovered)
		}
	}()

	params, err = s.process(users)
	return
}

// NewTask возвращает ссылку на новый экземпляр Task.
func NewTask(
	id Id,
	appId app.Id,
	from *internal.Time,
	to *internal.Time,
	process ProcessFunc,
) *Task {
	return &Task{
		AppId:   appId,
		Id:      id,
		From:    from,
		To:      to,
		process: process,
	}
}
