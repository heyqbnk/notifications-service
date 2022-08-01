package timezone

const (
	// MinTimezone - минимальное значение для часового пояса.
	MinTimezone = -720
	// MaxTimezone - максимальное значение для часового пояса.
	MaxTimezone = 840
)

// Timezone описывает часовой пояс, то есть то количество минут, которое
// необходимо прибавить ко времени по Гринвичу, чтобы получить локальное время.
// https://ru.wikipedia.org/wiki/Часовой_пояс
// Допустимые значения: от -720 до 840
type Timezone int

// IsValidTimezone возвращает true в случае, если переданное значение находится
// в диапазоне допустимых значений.
func IsValidTimezone(value int) bool {
	return -720 <= value && value <= 840
}

// CutTimezone проверяет, находится ли значение в допустимом диапазоне и в
// случае, если это не так, приравнивает к ближайшему значению допустимого
// диапазона.
func CutTimezone(tz Timezone) Timezone {
	if tz > MaxTimezone {
		return MaxTimezone
	}
	if tz < MinTimezone {
		return MinTimezone
	}
	return tz
}
