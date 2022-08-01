package service

import (
	"errors"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/wolframdeus/noitifications-service/internal/app"
	"github.com/wolframdeus/noitifications-service/internal/providers"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"time"
)

// SetAllowStatusForUser описывает функцию для изменения разрешения на
// отправку уведомлений пользователю.
type SetAllowStatusForUser func(appId app.Id, userId user.Id, allowed bool) error

// OnError описывает функцию, которая вызывается в случае возникновения
// внутренней ошибки.
type OnError func(error)

type NewOptions struct {
	// Интервал между итерациями сервиса, которые вызывают отправку уведомлений.
	TickInterval time.Duration
	OnError      OnError
}

type Service struct {
	provider providers.Provider
	// Интервал между итерациями сервиса, которые вызывают отправку уведомлений.
	tickInterval time.Duration
	// Список задач, выполняемых сервисом.
	tasks []task.Task
	// TODO: Обезопасить функцию.
	onError OnError
	// Экземпляр библиотеки для работы с API ВКонтакте.
	vk *api.VK
}

// AddTask добавляет новую задачу.
func (s *Service) AddTask(tasks ...task.Task) {
	s.tasks = append(s.tasks, tasks...)
}

// RegisterUser регистрирует пользователя в сервисе.
func (s *Service) RegisterUser(userId user.Id, tz timezone.Timezone) error {
	// FIXME: Реализовать безопасный вызов (safe).
	return s.provider.RegisterUser(userId, tz)
}

// Start выполняет запуск сервиса.
func (s *Service) Start() {
}

// Stop выполняет остановку сервиса.
func (s *Service) Stop() {
}

// SetAllowStatusForUser изменяет разрешение на отправку уведомлений
// пользователю.
// TODO: Возможно, стоит предоставить возможность выполнять upsert в случае,
//  если пользователь не существует.
func (s *Service) SetAllowStatusForUser(userId user.Id, appId app.Id, allowed bool) error {
	return s.safeSetAllowStatusForUser(userId, appId, allowed)
}

// UserExists возвращает true в случае, если пользователь зарегистрирован в
// сервисе.
func (s *Service) UserExists(userId user.Id) (bool, error) {
	// FIXME: Реализовать безопасный вызов (safe).
	return s.provider.UserExists(userId)
}

// New создаёт ссылку на новый экземпляр Service.
func New(provider providers.Provider, accessToken string, options NewOptions) (*Service, error) {
	if options.TickInterval == 0 {
		return nil, errors.New(`"TickInterval" не был указан`)
	}
	return &Service{
		provider: provider,
		onError:  options.OnError,
		vk:       api.NewVK(accessToken),
	}, nil
}
