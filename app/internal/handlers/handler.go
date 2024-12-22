package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"todo-list/app/internal/commandsDB"
	"todo-list/app/internal/models"
	"todo-list/app/internal/tasks"
)

// функция для обработки nextdate
func HandlerForNewDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, `{"error": "некорректная дата now"}`, http.StatusBadRequest)
		return
	}
	// Вызываем функцию NextDate
	nextDate, err := tasks.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(nextDate))
}

// функция для добавления задачи
func PostTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	decoder := json.NewDecoder(r.Body)

	var task = models.Remind{}
	// декодируем JSON
	if err := decoder.Decode(&task); err != nil {
		http.Error(w, `{"error":"Ошибка декодирования JSON"}`, http.StatusBadRequest)
		return
	}

	// Проверяем обязательное поле title
	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// начинаем проверку даты
	now := time.Now()
	today := now.Format("20060102")
	// если поле date не указано - берется сегодняшнее число
	if task.Date == "" || task.Date == "today" {
		task.Date = today
	} else {
		// парсим дату задачи
		parsedDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, `{"error":"Некорректный формат даты"}`, http.StatusBadRequest)
			return
		}
		// если дата меньше сегодняшнего числа, то
		if parsedDate.Before(time.Now()) {
			// правило не указано - дата сегодняшнее число
			if task.Repeat == "" {
				task.Date = today
				// правило указано - дата - та, что в правиле
			} else {
				nextDate, err := tasks.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
					return
				}
				if task.Date != today {
					task.Date = nextDate
				}
			}
		}

		// проверяем правило повторения в любом случае
		if task.Repeat != "" {
			_, err := tasks.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
				return
			}
		}
		id, err := commandsDB.InsertInDB(task)
		if err != nil {
			http.Error(w, `{"error":"Ошибка записи в БД"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": id,
		})
	}
}

// отображение задач
func GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")

	// ограничение на кол-во возвращаемых задач
	const limit = 50

	tasks, err := commandsDB.FindInDB(search, limit)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// преобразуем задачи в формат, ожидаемый тестами
	var tasksForResponse []map[string]string
	for _, task := range tasks {
		tasksForResponse = append(tasksForResponse, map[string]string{
			"id":      fmt.Sprintf("%d", task.ID),
			"date":    task.Date,
			"title":   task.Title,
			"comment": task.Comment,
			"repeat":  task.Repeat,
		})
	}

	// если пустой, то возвращаем слайс
	if tasksForResponse == nil {
		tasksForResponse = []map[string]string{}
	}

	// формируем JSON-ответ
	response := map[string]interface{}{
		"tasks": tasksForResponse,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, `{"error":"Ошибка формирования JSON-ответа"}`, http.StatusInternalServerError)
		return
	}
}
