package commandsDB

import (
	"errors"
	"time"
	"todo-list/app/config"
	"todo-list/app/internal/models"
)

func InsertInDB(task models.Remind) (uint64, error) {
	query := "INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)"
	res, err := config.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	// получаем ID созданной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func FindInDB(search string, limit int) ([]models.Remind, error) {
	var query string
	var args []interface{}

	// Если параметр search пустой, возвращаем все задачи
	if search == "" {
		query = "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?"
		args = append(args, limit)
	} else {
		if parsedDate, err := time.Parse("02.01.2006", search); err == nil {
			// если дата, используем сравнение по дате
			query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?"
			args = append(args, parsedDate.Format("20060102"), limit)
		} else {
			// если подстрока - в title и comment
			likePattern := "%" + search + "%"
			query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?"
			args = append(args, likePattern, likePattern, limit)
		}
	}

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, errors.New("ошибка выполнения запроса к базе данных")
	}
	defer rows.Close()

	var tasks []models.Remind
	for rows.Next() {
		var task models.Remind
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, errors.New("ошибка сканирования данных из базы")
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("ошибка постобработки данных из базы")
	}
	return tasks, nil
}
