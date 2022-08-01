package timezone

// Range описывает промежуток интервалов.
type Range struct {
	// Начало промежутка.
	From Timezone
	// Конец промежутка.
	To Timezone
}

// ContainsTimezone вернёт true в случае, если переданный часовой пояс находится
// в указанном диапазоне.
func (r *Range) ContainsTimezone(value Timezone) bool {
	return r.From <= value && value <= r.To
}

// NewRange создает ссылку на новый экземпляр Range.
func NewRange(from Timezone, to Timezone) *Range {
	return &Range{From: from, To: to}
}
