package errors

import (
	"github.com/wolframdeus/noitifications-service/internal/app"
	"github.com/wolframdeus/noitifications-service/internal/task"
)

type TaskError struct {
	// Идентификатор приложения владельца уведомления.
	AppId app.Id
	// Идентификатор задачи.
	TaskId task.Id
	// Оригинальная выброшенная ошибка.
	Original error
}

// NewTaskError возвращает ссылку на новый экземпляр TaskError.
func NewTaskError(appId app.Id, taskId task.Id, err error) *TaskError {
	return &TaskError{
		AppId:    appId,
		TaskId:   taskId,
		Original: err,
	}
}
