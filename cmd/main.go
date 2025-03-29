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
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables:", err)
	} else {
		log.Println("Successfully loaded .env file")
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is not set. Please set it in .env file or environment variables.")
	} else {
		log.Println("GEMINI_API_KEY is set")
	}

	r := gin.Default()

	r.POST("/upload", handlers.UploadDocumentGin) // Новая функция
	r.POST("/add", handlers.AddDocuments)
	r.POST("/query", handlers.Query)

	defer services.CloseConnection()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
