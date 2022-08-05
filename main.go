package main

import (
	"github.com/getsentry/sentry-go"
	"github.com/wolframdeus/noitifications-service/internal"
	"github.com/wolframdeus/noitifications-service/internal/appid"
	"github.com/wolframdeus/noitifications-service/internal/errors"
	"github.com/wolframdeus/noitifications-service/internal/notification"
	"github.com/wolframdeus/noitifications-service/internal/providers/mongodb"
	"github.com/wolframdeus/noitifications-service/internal/service"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/taskid"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"log"
	"time"
)

const (
	HealthAppId       appid.Id  = 7865682
	HealthSomeTaskId1 taskid.Id = 1
)

func main() {
	provider, err := mongodb.New("localhost", 27017, "notifications-service", 100)
	if err != nil {
		panic(err)
	}

	// Создаём новый сервис.
	// FIXME: access token
	s, err := service.New(provider, "accessToken", service.NewOptions{
		TickInterval: 10 * time.Minute,
		SentryOptions: &sentry.ClientOptions{
			Dsn:              "https://792ef54fbc6e40eaaa6123514e06948a@o992980.ingest.sentry.io/6625183",
			Debug:            true,
			TracesSampleRate: 1.0,
		},
	})
	if err != nil {
		panic(err)
	}

	// Добавляем новую задачу.
	s.AddTask(
		*task.NewTask(
			HealthSomeTaskId1,
			HealthAppId,
			internal.NewTime(00, 00),
			internal.NewTime(2, 00),
			func(users []user.User) ([]notification.Params, *errors.TaskError) {
				res := make([]notification.Params, len(users))

				for i, u := range users {
					res[i] = notification.Params{UserId: u.Id, Message: "Ты попал в задачу!"}
				}
				return res, nil
			},
		),
	)

	if err := s.SetAllowStatusForUser(898, 521, true, nil); err != nil {
		log.Println(err)
	}
	s.Cleanup()
}
