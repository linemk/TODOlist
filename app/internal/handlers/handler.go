package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"
	"todo-list-project/app/internal/commandsDB"
	"todo-list-project/app/internal/models"
	"todo-list-project/app/internal/tasksRules"
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
	nextDate, err := tasksRules.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(nextDate))
}

// добавление задачи
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
				nextDate, err := tasksRules.NextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
					return
				}
				if task.Date != today {
					task.Date = nextDate
				}
			}
		}
	}
	// проверяем правило повторения в любом случае
	if task.Repeat != "" {
		_, err := tasksRules.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
			return
		}
	}
	id, err := commandsDB.InsertInDB(task)
	idForResponse := strconv.Itoa(int(id))

	if err != nil {
		http.Error(w, `{"error":"Ошибка записи в БД"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{
		"id": idForResponse,
	})
}

// отображение задач + поиск конкретной задачи
func GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	search := r.URL.Query().Get("search")

	// ограничение на кол-во возвращаемых задач
	const limit = 50

	tasks, err := commandsDB.FindinDb(search, limit)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// преобразуем задачи в формат, ожидаемый тестами
	var tasksForResponse []map[string]string
	for _, task := range tasks {
		tasksForResponse = append(tasksForResponse, map[string]string{
			"id":      task.ID,
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

// получение одной задачи по ID
func GetOneTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"Некорректный идентификатор"}`, http.StatusBadRequest)
		return
	}
	// получаем задачу
	task, err := commandsDB.GetTaskByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	response := map[string]string{
		"id":      task.ID,
		"date":    task.Date,
		"title":   task.Title,
		"comment": task.Comment,
		"repeat":  task.Repeat,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, `{"error":"Ошибка формирования JSON-ответа"}`, http.StatusInternalServerError)
		return
	}
}

// обновление задачи
func PutTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var task models.Remind
	if err := decoder.Decode(&task); err != nil {
		http.Error(w, `{"error":"Ошибка декодирования JSON"}`, http.StatusBadRequest)
		return
	}

	// проверяем наличие поля id
	if task.ID == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	// проверяем также поле title
	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	// проверяем дату
	now := time.Now()
	today := now.Format("20060102")
	if task.Date == "" || task.Date == "today" {
		task.Date = today
	} else {
		parsedDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, `{"error":"Некорректный формат даты"}`, http.StatusBadRequest)
			return
		}
		if parsedDate.Before(time.Now()) {
			if task.Repeat == "" {
				task.Repeat = today
			} else {
				nextDate, err := tasksRules.NextDate(now, task.Date, task.Repeat)
				if err != nil {
					http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
					return
				}
				if task.Date != today {
					task.Date = nextDate
				}
			}
		}
	}
	// проверяем правило повторения
	if task.Repeat != "" {
		_, err := tasksRules.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Некорректное правило повторения"}`, http.StatusBadRequest)
			return
		}
	}

	// Обновляем задачу в базе данных
	err := commandsDB.UpdateTask(task)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Ошибка обновления задачи"}`, http.StatusInternalServerError)
		}
		return
	}
	// Возвращаем пустой JSON-объект
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// отметка о выполнении задачи
func DoneTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	idStr := chi.URLParam(r, "id") // Извлечение параметра из пути
	if idStr == "" {               // Если параметр пустой, пробуем из строки запроса
		idStr = r.URL.Query().Get("id")
	}

	if idStr == "" { // Если параметр всё ещё пустой, возвращаем ошибку
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"Некорректный идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	// получаем задачу из базы данных
	task, err := commandsDB.GetTaskByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Ошибка получения задачи из БД"}`, http.StatusInternalServerError)
		}
		return
	}

	// проверка на пустоту продолжения задачи
	if task.Repeat == "" {
		err := commandsDB.DeleteTaskByID(id)
		if err != nil {
			http.Error(w, `{"error":"Ошибка удаления задачи"}`, http.StatusInternalServerError)
			return
		}
	} else {
		now := time.Now()
		nextDate, err := tasksRules.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Ошибка расчёта следующей даты повторения"}`, http.StatusInternalServerError)
			return
		}

		// Обновляем дату задачи в базе данных
		err = commandsDB.UpdateTaskDate(uint64(id), nextDate)
		if err != nil {
			http.Error(w, `{"error":"Ошибка обновления задачи"}`, http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

// удаление задачи
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"Некорректный идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	err = commandsDB.DeleteTaskByID(id)
	if err != nil {
		http.Error(w, `{"error":"Ошибка удаления задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}
