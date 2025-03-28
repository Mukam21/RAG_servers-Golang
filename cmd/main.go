package main

import (
	"log"
	"os"

	"github.com/Mukam21/RAG_server-Golang/pkg/handlers"
	"github.com/Mukam21/RAG_server-Golang/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения из .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables:", err)
	} else {
		log.Println("Successfully loaded .env file")
	}

	// Проверяем, загрузилась ли переменная GEMINI_API_KEY
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is not set. Please set it in .env file or environment variables.")
	} else {
		log.Println("GEMINI_API_KEY is set:", geminiAPIKey)
	}

	r := gin.Default()

	// Маршруты
	r.POST("/add", handlers.AddDocuments)
	r.POST("/query", handlers.Query)

	// Запуск сервера
	defer services.CloseConnection()
	r.Run(":8080")
}
