package config

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadEnviroment() {
	// загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}
}
