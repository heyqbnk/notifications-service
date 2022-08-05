package errors

import (
	"github.com/wolframdeus/noitifications-service/internal/appid"
	"github.com/wolframdeus/noitifications-service/internal/taskid"
)

type TaskError struct {
	// Идентификатор приложения владельца уведомления.
	AppId appid.Id
	// Идентификатор задачи.
	TaskId taskid.Id
	// Оригинальная выброшенная ошибка.
	Original error
}

// NewTaskError возвращает ссылку на новый экземпляр TaskError.
func NewTaskError(appId appid.Id, taskId taskid.Id, err error) *TaskError {
	return &TaskError{
		AppId:    appId,
		TaskId:   taskId,
		Original: err,
	}
}
