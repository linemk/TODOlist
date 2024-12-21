package tasks

import (
	"errors"
	"strconv"
	"time"
)

// обработчик дня
func HandleDailyRepeat(now time.Time, taskDate time.Time, rules []string) (string, error) {
	// проверяем правильность структуры правила
	if len(rules) != 2 {
		return "", errors.New("неверный формат правила повторения для дней")
	}

	// преобразуем количество дней в число
	days, err := strconv.Atoi(rules[1])
	if err != nil || days <= 0 || days > 400 {
		return "", errors.New("некорректный интервал для правила повторения дней")
	}
	if taskDate.After(now) {
		taskDate = taskDate.AddDate(0, 0, days)
	} else {
		// Добавляем дни до тех пор, пока `taskDate` не станет больше `now`
		for !taskDate.After(now) {
			taskDate = taskDate.AddDate(0, 0, days)
		}
	}
	// Возвращаем следующую дату в формате YYYYMMDD
	return taskDate.Format("20060102"), nil
}
