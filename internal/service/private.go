package service

import (
	"github.com/getsentry/sentry-go"
	customerror "github.com/wolframdeus/noitifications-service/internal/errors"
	"strconv"
)

type CaptureOptions struct {
	// Список тегов.
	Tags map[string]string
	// Список контекстов.
	Contexts map[string]interface{}
	// Уровень сообщения.
	Level sentry.Level
}

// Логирует ошибку, возникшую в сервисе.
func (s *Service) captureServiceError(
	err *customerror.ServiceError,
	options *CaptureOptions,
) {
	s.sentryHub.WithScope(func(scope *sentry.Scope) {
		if options != nil {
			scope.SetTags(options.Tags)
			scope.SetContexts(options.Contexts)
			scope.SetLevel(options.Level)
		}
		s.sentryHub.CaptureException(err.Original)
	})
}

// Логирует ошибку, возникшую в задаче.
func (s *Service) captureTaskError(
	err *customerror.TaskError,
	options *CaptureOptions,
) {
	s.sentryHub.WithScope(func(scope *sentry.Scope) {
		if options != nil {
			scope.SetTags(options.Tags)
			scope.SetContexts(options.Contexts)
			scope.SetLevel(options.Level)

			scope.SetTag("task-id", strconv.Itoa(int(err.TaskId)))
			scope.SetTag("app-id", strconv.Itoa(int(err.AppId)))
		}
		s.sentryHub.CaptureException(err.Original)
	})
}
