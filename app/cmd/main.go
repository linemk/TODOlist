package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"todo-list-project/app/config"
	"todo-list-project/app/internal/handlers"
)

var (
	portLocal = "7540"
	webDir    = "./web"
)

func main() {
	config.LoadEnviroment()
	config.MakeDB()
	defer config.CloseDB()

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = portLocal
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/*", http.FileServer(http.Dir(webDir)))

	// обработчик для авторизации
	r.Post("/api/signin", handlers.SignInHandler)

	// защищенные маршруты будут прогоняться через токен
	r.Route("/api", func(r chi.Router) {
		// добавляем middleware
		r.Use(handlers.AuthMiddleware)

		// вычисление следующей даты
		r.Get("/nextdate", handlers.HandlerForNewDate)
		// добавление следующей задачи
		r.Post("/task", handlers.PostTask)
		// отображение задач
		r.Get("/tasks", handlers.GetTasks)
		// получение одной задачи
		r.Get("/task", handlers.GetOneTask)
		// изменение одной задачи
		r.Put("/task", handlers.PutTask)
		// выполнение задачи
		r.Post("/task/done", handlers.DoneTask)
		// удаление задачи
		r.Delete("/task", handlers.DeleteTask)
	})

	// запуск сервера
	log.Printf("Сервер запущен на http://localhost:%s/", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
