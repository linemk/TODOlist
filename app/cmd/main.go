package main

import (
	"log"
	"net/http"
	"os"
	config "todo-list/app/config"
)

var (
	portLocal = "7540"
	webDir    = "./web"
)

func main() {
	config.LoadEnviroment()

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = portLocal
	}

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Запускаем сервер
	log.Printf("Сервер запущен на http://localhost:%s/", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
