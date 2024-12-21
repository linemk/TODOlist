package tasks

import (
	"errors"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	// откуда начинается отсчет
	taskDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.New("некорректный формат даты")
	}

	if repeat == "" {
		return "", errors.New("правило повторения не указано")
	}

	rules := strings.Split(repeat, " ")
	switch rules[0] {
	// правило для дней
	case "d":
		return HandleDailyRepeat(now, taskDate, rules)
	case "y":
		return HandleYearRepeat(now, taskDate, rules)
	default:
		return "", errors.New("правило повторения не поддерживается")
	}
}
