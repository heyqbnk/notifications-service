package internal

import (
	"github.com/wolframdeus/noitifications-service/internal/timezone"
	"time"
)

// Time описывает структуру времени.
type Time struct {
	// Часы.
	Hours byte
	// Минуты.
	Minutes byte
}

// GetTimezones возвращает массив часовых поясов, в которых на данный момент
// указанное время. В качестве результата возвращается массив из 1 или двух
// значений отсортированных по возрастанию.
func (t *Time) GetTimezones() []timezone.Timezone {
	// Получаем текущую дату по Гринвичу.
	now := time.Now().UTC()

	// Получаем предполагаемую будущую дату с установленным временем из
	// текущего экземпляра.
	futureDate := t.insertInto(now)

	// Создаём ссылку для сохранения даты на предыдущий день.
	pastDate := time.Now()

	// Переходим к проверке тех дат, с которыми будем далее работать. Вычисляем
	// будущую и прошлую даты.
	if futureDate.After(now) {
		pastDate = t.insertInto(now.Add(-24 * time.Hour))
	} else {
		pastDate, futureDate = futureDate, t.insertInto(now.Add(24*time.Hour))
	}

	// Находим значения часовых поясов.
	futureTz := futureDate.Sub(now).Minutes()
	pastTz := pastDate.Sub(now).Minutes()

	res := make([]timezone.Timezone, 0, 2)

	// Проверяем каждый из часовых поясов и добавляем в результирующий массив.
	if timezone.IsValidTimezone(int(pastTz)) {
		res = append(res, timezone.Timezone(pastTz))
	}
	if timezone.IsValidTimezone(int(futureTz)) {
		res = append(res, timezone.Timezone(futureTz))
	}
	return res
}

// Вставляет в указанный экземпляр даты текущие значения часов и минут.
func (t *Time) insertInto(date time.Time) time.Time {
	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		int(t.Hours),
		int(t.Minutes),
		date.Second(),
		date.Nanosecond(),
		date.Location(),
	)
}

// NewTime возвращает ссылку на новый экземпляр Time.
func NewTime(h byte, m byte) *Time {
	return &Time{Hours: h, Minutes: m}
}
