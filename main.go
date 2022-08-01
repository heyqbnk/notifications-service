package main

import (
	"fmt"
	"github.com/wolframdeus/noitifications-service/internal"
	"github.com/wolframdeus/noitifications-service/internal/app"
	"github.com/wolframdeus/noitifications-service/internal/notification"
	"github.com/wolframdeus/noitifications-service/internal/providers/mongodb"
	"github.com/wolframdeus/noitifications-service/internal/service"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"time"
)

const (
	HealthAppId       app.Id  = 7865682
	HealthSomeTaskId1 task.Id = 1
)

func main() {
	provider, err := mongodb.New("localhost", 27017, "notifications-service")
	if err != nil {
		panic(err)
	}

	// Создаём новый сервис.
	s, err := service.New(provider, accessToken, service.NewOptions{
		OnError: func(err error) {
			fmt.Printf("Erorr occurred: %s", err.Error())
		},
		TickInterval: 10 * time.Minute,
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
			func(users []user.User) ([]notification.Params, error) {
				res := make([]notification.Params, len(users))

				for i, u := range users {
					res[i] = notification.Params{UserId: u.Id, Message: "Ты попал в задачу!"}
				}
				return res, nil
			},
		),
	)

	s.RunIteration()
}
