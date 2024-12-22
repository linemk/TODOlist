package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"todo-list/app/config"
	"todo-list/app/internal/handlers"
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
	// API-маршрут для вычисления следующей даты
	r.Get("/api/nextdate", handlers.HandlerForNewDate)
	// добавление следующей задачи
	r.Post("/api/task", handlers.PostTask)
	// отображение задач
	r.Get("/api/tasks", handlers.GetTasks)

	// Запуск сервера
	log.Printf("Сервер запущен на http://localhost:%s/", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
