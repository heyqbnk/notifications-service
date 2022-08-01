package mongodb

import (
	"context"
	"fmt"
	"github.com/wolframdeus/noitifications-service/internal/app"
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
) (*providers.GetUsersByTimezonesResult, error) {
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
		// TODO: Возвращать общую ошибку.
		return nil, err
	}

	var users []user.User

	for cur.Next(context.TODO()) {
		var u User

		if err := cur.Decode(&u); err != nil {
			// TODO: Возвращать общую ошибку.
			return nil, err
		}
		// TODO: Вынести инициализацию user.User из User в отдельный метод.
		users = append(users, user.User{
			Id:       user.Id(u.Id),
			Timezone: timezone.Timezone(u.Timezone),
		})
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
) error {
	// TODO: Уйти от этого подхода в сторону работы со структурой User.
	path := fmt.Sprintf("apps.%d.areNotificationsEnabled", appId)
	res, err := p.getUsersCollection().UpdateByID(
		context.Background(),
		userId,
		bson.D{{"$set", bson.D{{path, allowed}}}},
	)
	if err != nil {
		// TODO: Возвращать общую ошибку.
		return err
	}
	if res.MatchedCount == 0 {
		return providers.ErrUserDoesNotExist
	}
	return nil
}

func (p *Provider) SaveNotificationDate(
	userIds []user.Id,
	appId app.Id,
	taskId task.Id,
	date time.Time,
) error {
	path := fmt.Sprintf("apps.%d.tasks.%d.history", appId, taskId)
	res, err := p.
		getUsersCollection().
		UpdateMany(
			context.Background(),
			bson.M{"_id": bson.M{"$in": userIds}},
			bson.M{
				"$push": bson.D{{path, bson.M{
					"$each":     []time.Time{date},
					"$position": 0,
					"$slice":    15, // TODO: Вынести в опцию?
				}}},
			},
		)
	if err != nil {
		return err
	}
	fmt.Println(path, res.MatchedCount, userIds)
	return nil
}

func (p *Provider) UserExists(userId user.Id) (bool, error) {
	var limit int64 = 1
	count, err := p.
		getUsersCollection().
		CountDocuments(
			context.Background(),
			bson.D{{"_id", userId}},
			&options.CountOptions{Limit: &limit},
		)
	if err != nil {
		// TODO: Возвращать общую ошибку.
		return false, err
	}
	return count > 0, nil
}

func (p *Provider) RegisterUser(userId user.Id, tz timezone.Timezone) error {
	u := NewUser(UserId(userId), Apps{}, int(tz))
	_, err := p.getUsersCollection().InsertOne(context.Background(), u)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return providers.ErrUserAlreadyExists
		}
		// TODO: Возвращать общую ошибку.
		return err
	}
	return nil
}

// Возвращает коллекцию пользователей.
func (p *Provider) getUsersCollection() *mongo.Collection {
	return p.client.Database(p.db).Collection("users")
}

// New возвращает новый экземпляр драйвера для работы с MongoDB.
func New(host string, port uint, db string) (providers.Provider, error) {
	connString := fmt.Sprintf("mongodb://%s:%d", host, port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connString))
	if err != nil {
		return nil, err
	}

	return &Provider{
		client: client,
		db:     db,
		// TODO: Вынести в опцию конструктора.
		getUsersByTimezonesLimit: 10000,
	}, nil
}
