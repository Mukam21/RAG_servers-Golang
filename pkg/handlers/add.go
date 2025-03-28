package handlers

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/Mukam21/RAG_server-Golang/pkg/services"
	"github.com/gin-gonic/gin"
)

type AddRequest struct {
	Documents []struct {
		Text string `json:"text" binding:"required"`
	} `json:"documents" binding:"required,dive"`
}

func AddDocuments(c *gin.Context) {
	var req AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if len(req.Documents) == 0 {
		log.Println("Documents list is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "documents list cannot be empty"})
		return
	}

	// Канал для ошибок
	errChan := make(chan error, len(req.Documents))
	var wg sync.WaitGroup

	for _, doc := range req.Documents {
		trimmedText := strings.TrimSpace(doc.Text)
		if len(trimmedText) < 5 {
			log.Println("Document text too short:", trimmedText)
			c.JSON(http.StatusBadRequest, gin.H{"error": "document text must be at least 5 characters long"})
			return
		}

		wg.Add(1)
		go func(text string) {
			defer wg.Done()

			// Получаем эмбеддинг
			log.Println("Fetching embedding for text:", text)
			embedding, err := services.GetEmbedding(text)
			if err != nil {
				log.Println("Failed to get embedding:", err)
				errChan <- err
				return
			}

			// Сохраняем в PostgreSQL
			log.Println("Saving document to PostgreSQL:", text)
			if err := services.AddDocument(text, embedding); err != nil {
				log.Println("Failed to save document to PostgreSQL:", err)
				errChan <- err
				return
			}
		}(trimmedText)
	}

	// Ждем завершения всех горутин
	wg.Wait()
	close(errChan)

	// Проверяем ошибки
	for err := range errChan {
		if err != nil {
			log.Println("Error processing documents:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process documents: " + err.Error()})
			return
		}
	}

	log.Println("Successfully added", len(req.Documents), "documents")
	c.JSON(http.StatusOK, gin.H{"message": "documents added successfully"})
}
