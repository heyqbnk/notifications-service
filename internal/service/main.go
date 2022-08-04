package service

import (
	"errors"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/getsentry/sentry-go"
	"github.com/wolframdeus/noitifications-service/internal/app"
	customerror "github.com/wolframdeus/noitifications-service/internal/errors"
	"github.com/wolframdeus/noitifications-service/internal/providers"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"time"
)

// SetAllowStatusForUser описывает функцию для изменения разрешения на
// отправку уведомлений пользователю.
type SetAllowStatusForUser func(appId app.Id, userId user.Id, allowed bool) error

type NewOptions struct {
	// Интервал между итерациями сервиса, которые вызывают отправку уведомлений.
	TickInterval time.Duration
	// Список опций, которые далее передаются для инициализации Sentry Hub.
	SentryOptions *sentry.ClientOptions
}

type Service struct {
	provider providers.Provider
	// Интервал между итерациями сервиса, которые вызывают отправку уведомлений.
	tickInterval time.Duration
	// Список задач, выполняемых сервисом.
	tasks []task.Task
	// Экземпляр библиотеки для работы с API ВКонтакте.
	vk *api.VK
	// Hub Sentry для логирования ошибок.
	sentryHub *sentry.Hub
}

// AddTask добавляет новую задачу.
func (s *Service) AddTask(tasks ...task.Task) {
	s.tasks = append(s.tasks, tasks...)
}

// Start выполняет запуск сервиса.
func (s *Service) Start() {
	// TODO: Implement.
}

// Stop выполняет остановку сервиса.
func (s *Service) Stop() {
	// TODO: Implement.
}

// SetAllowStatusForUser изменяет разрешение на отправку уведомлений
// пользователю.
// TODO: Возможно, стоит предоставить возможность выполнять upsert в случае,
//  если пользователь не существует.
func (s *Service) SetAllowStatusForUser(
	userId user.Id,
	appId app.Id,
	allowed bool,
) *customerror.ServiceError {
	return s.safeSetAllowStatusForUser(userId, appId, allowed)
}

// New создаёт ссылку на новый экземпляр Service.
func New(
	provider providers.Provider,
	accessToken string,
	options NewOptions,
) (*Service, error) {
	if options.TickInterval == 0 {
		return nil, errors.New(`"TickInterval" не был указан`)
	}
	sentryHub := sentry.CurrentHub().Clone()

	// Если указаны опции инициализации Sentry-клиента, используем их.
	if options.SentryOptions != nil {
		client, err := sentry.NewClient(*options.SentryOptions)
		if err != nil {
			return nil, err
		}
		sentryHub.BindClient(client)
	}
	return &Service{
		provider:  provider,
		vk:        api.NewVK(accessToken),
		sentryHub: sentryHub,
	}, nil
}
