package tasks

import (
	"errors"
	"time"
)

func HandleYearRepeat(now time.Time, taskDate time.Time, rules []string) (string, error) {
	if len(rules) != 1 {
		return "", errors.New("неверный формат правила повторения для года")
	}
	if taskDate.After(now) {
		taskDate = taskDate.AddDate(1, 0, 0)
	} else {
		// Добавляем дни до тех пор, пока `taskDate` не станет больше `now`
		for !taskDate.After(now) {
			taskDate = taskDate.AddDate(1, 0, 0)
		}
	}
	// Возвращаем следующую дату в формате YYYYMMDD
	return taskDate.Format("20060102"), nil
}
