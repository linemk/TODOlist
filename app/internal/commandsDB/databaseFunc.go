package commandsDB

import (
	"errors"
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

func FindInDB(limit int) ([]models.Remind, error) {
	query := "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?"

	rows, err := config.DB.Query(query, limit)
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
