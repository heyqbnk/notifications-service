package mongodb

import (
	"context"
	"fmt"
	"github.com/wolframdeus/noitifications-service/internal/app"
	customerror "github.com/wolframdeus/noitifications-service/internal/errors"
	"github.com/wolframdeus/noitifications-service/internal/notification"
	"github.com/wolframdeus/noitifications-service/internal/providers"
	"github.com/wolframdeus/noitifications-service/internal/task"
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"github.com/wolframdeus/noitifications-service/internal/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Provider struct {
	// Клиент MongoDB.
	client *mongo.Client
	// Наименование БД источника данных.
	db string
	// Максимальное количество пользователей, которое может быть возвращено
	// методом GetUsersByTimezones.
	getUsersByTimezonesLimit int64
}

func (p *Provider) GetUsersByTimezones(
	tz []timezone.Range,
	cursor user.Id,
) (*providers.GetUsersByTimezonesResult, *customerror.ServiceError) {
	if len(tz) == 0 {
		return providers.NewGetUsersByTimezonesResult(0, nil, false), nil
	}
	// Составляем условие запроса для MongoDB.
	orQuery := make([]bson.M, len(tz))
	for i, t := range tz {
		orQuery[i] = bson.M{"timezone": bson.M{"$gte": t.From, "$lte": t.To}}
	}

	cur, err := p.
		getUsersCollection().
		Find(
			context.Background(),
			bson.M{
				"_id": bson.M{"$gt": cursor},
				"$or": orQuery,
			},
			options.
				Find().
				SetLimit(p.getUsersByTimezonesLimit+1).
				SetSort(bson.D{{"_id", 1}}),
		)
	if err != nil {
		return nil, customerror.NewServiceError(err)
	}

	var users []user.User

	for cur.Next(context.TODO()) {
		var u User

		if err := cur.Decode(&u); err != nil {
			return nil, customerror.NewServiceError(err)
		}
		users = append(users, *u.ToCommon())
	}

	// Пользователей нет, возвращаем стандартный ответ.
	if len(users) == 0 {
		return providers.NewGetUsersByTimezonesResult(0, nil, false), nil
	}

	hasMore := false
	// Получили больше пользователей чем разрешено. Это означает, что есть
	// больше пользователей удовлетворяющих условию.
	if int64(len(users)) > p.getUsersByTimezonesLimit {
		// Отсекаем последнего пользователя.
		users = users[0 : len(users)-1]
		hasMore = true
	}
	return providers.NewGetUsersByTimezonesResult(users[len(users)-1].Id, users, hasMore), nil
}

func (p *Provider) SetAllowStatusForUser(
	userId user.Id,
	appId app.Id,
	allowed bool,
) *customerror.ServiceError {
	// TODO: Уйти от этого подхода в сторону работы со структурой User.
	path := fmt.Sprintf("apps.%d.areNotificationsEnabled", appId)
	res, err := p.getUsersCollection().UpdateByID(
		context.Background(),
		userId,
		bson.D{{"$set", bson.D{{path, allowed}}}},
	)
	if err != nil {
		return customerror.NewServiceError(err)
	}
	if res.MatchedCount == 0 {
		return customerror.NewServiceError(providers.ErrUserDoesNotExist)
	}
	return nil
}

func (p *Provider) SaveSendResult(
	results *notification.SendResult,
	appId app.Id,
	taskId task.Id,
	date time.Time,
) *customerror.ServiceError {
	// TODO: Возможно это сделать в виде агрегации?
	// TODO: Возвращать этот массив ошибок.
	// TODO: Выполнять все обновления в отдельных горутинах?
	var errs []customerror.ServiceError

	// Обновляем пользователей, которым удалось отправить уведомление.
	if results.Success != nil && len(results.Success) > 0 {
		path := fmt.Sprintf("apps.%d.tasks.%d.history", appId, taskId)
		_, err := p.
			getUsersCollection().
			UpdateMany(
				context.Background(),
				bson.M{"_id": bson.M{"$in": results.Success}},
				bson.M{
					"$push": bson.D{{path, bson.M{
						"$each":     []time.Time{date},
						"$position": 0,
						"$slice":    15, // TODO: Вынести в опцию?
					}}},
				},
			)
		if err != nil {
			errs = append(errs, *customerror.NewServiceError(err))
		}
	}

	// Обновляем пользователей, уведомления которым запрещены.
	if results.NotificationsDisabled != nil && len(results.NotificationsDisabled) > 0 {
		path := fmt.Sprintf("apps.%d.areNotificationsEnabled", appId)
		_, err := p.
			getUsersCollection().
			UpdateMany(
				context.Background(),
				bson.M{"_id": bson.M{"$in": results.Success}},
				bson.D{{path, false}},
			)
		if err != nil {
			errs = append(errs, *customerror.NewServiceError(err))
		}
	}

	// TODO: Обновить пользователей, у которых кулдаун на отправку уведомления.

	return nil
}

// Возвращает коллекцию пользователей.
func (p *Provider) getUsersCollection() *mongo.Collection {
	return p.client.Database(p.db).Collection("users")
}

// New возвращает новый экземпляр драйвера для работы с MongoDB.
func New(
	host string,
	port uint,
	db string,
	getUsersByTimezonesLimit int64,
) (providers.Provider, error) {
	connString := fmt.Sprintf("mongodb://%s:%d", host, port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connString))
	if err != nil {
		return nil, err
	}

	return &Provider{
		client:                   client,
		db:                       db,
		getUsersByTimezonesLimit: getUsersByTimezonesLimit,
	}, nil
}
