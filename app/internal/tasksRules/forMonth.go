package tasksRules

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

// извлекаем правила
func parseMonthDays(days string) ([]int, error) {
	daysParts := strings.Split(days, ",")
	var dayInt []int
	// обозначаем правила
	for _, part := range daysParts {
		day, err := strconv.Atoi(part)
		if err != nil || day == 0 || day < -31 || day > 31 {
			return nil, errors.New("некорректный день месяца в правиле повторения")
		}
		dayInt = append(dayInt, day)
	}
	return dayInt, nil
}

// извлекаем месяца
func parseMonth(monthStr string) ([]int, error) {
	monthsParts := strings.Split(monthStr, ",")
	var months []int
	for _, part := range monthsParts {
		month, err := strconv.Atoi(part)
		if err != nil || month < 1 || month > 12 {
			return nil, errors.New("некорректный месяц в правиле повторения")
		}
		months = append(months, month)
	}
	return months, nil
}

// проверка - входит ли месяц в список разрешенных
func isMonthAllowed(currentMonth int, months []int) bool {
	for _, month := range months {
		if currentMonth == month {
			return true
		}
	}
	return false
}

// Вычисляет целевую дату для указанного дня месяца
func calculateTargetDate(now, taskDate time.Time, days []int, allowedMonths []int) time.Time {
	// год и месяц от какого числа мы идем
	year, month, _ := taskDate.Date()
	location := taskDate.Location()

	// сортируем массив дней для обработки от меньшего к большему (например, от -1 до 5)
	sort.Ints(days)
	// переменная для хранения ближайшей даты
	var nearestDate *time.Time

	for {
		// если есть список разрешённых месяцев, проверяем, подходит ли текущий месяц
		if len(allowedMonths) > 0 && !isMonthAllowed(int(month), allowedMonths) {
			month++
			if month > 12 {
				month = 1
				year++
			}
			continue
		}

		// Обрабатываем массив дней
		for _, day := range days {
			// итоговая дата
			var targetDate time.Time

			// для отрицательных дней
			if day < 0 {
				// Последний день текущего месяца
				firstDayNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, location)
				lastDayThisMonth := firstDayNextMonth.AddDate(0, 0, -1)
				targetDate = lastDayThisMonth.AddDate(0, 0, day+1) // Например, -1 = последний день
			} else {
				// для положительных дней
				targetDate = time.Date(year, month, day, 0, 0, 0, 0, location)
				// проверяем, существует ли такой день в месяце
				if targetDate.Month() != time.Month(month) {
					continue
				}
			}

			// Если targetDate больше now, проверяем, является ли она ближайшей
			if targetDate.After(now) {
				if nearestDate == nil || targetDate.Before(*nearestDate) {
					// сверяемся, раньше или позже ближайшая дата с таргетом
					nearestDate = &targetDate
				}
			}
		}

		// Если ближайшая дата найдена, возвращаем её
		if nearestDate != nil {
			return *nearestDate
		}

		// Если ни один день не подошёл, переходим к следующему месяцу
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
}

// главный обработчик
func HandleMonthRepeat(now time.Time, taskDate time.Time, rules []string) (string, error) {
	if len(rules) < 2 || len(rules) > 3 {
		return "", errors.New("неверный формат правила повторения для месяца")
	}

	// парсим дни месяца
	days, err := parseMonthDays(rules[1])
	if err != nil {
		return "", err
	}

	// парсим сами месяцы, если они указаны
	var months []int
	if len(rules) == 3 {
		months, err = parseMonth(rules[2])
		if err != nil {
			return "", err
		}
	}

	// Проверка корректности отрицательных дней
	for _, day := range days {
		if day < -2 || day == 0 || day > 31 {
			return "", errors.New("некорректный день в правиле повторения")
		}
	}

	for {
		// проверяем, входит ли месяц в список допустимых (если он указан)
		if len(months) > 0 && !isMonthAllowed(int(taskDate.Month()), months) {
			taskDate = taskDate.AddDate(0, 1, 0) // переходим к следующему месяцу
			continue
		}
		// проверяем дни месяца
		targetDate := calculateTargetDate(now, taskDate, days, months)
		if targetDate.After(now) {
			return targetDate.Format("20060102"), nil
		}
		// Переходим к следующему месяцу
		taskDate = taskDate.AddDate(0, 1, 0)
	}
}
