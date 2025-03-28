package handlers

import (
	"net/http"
	"strings"

	"github.com/Mukam21/RAG_server-Golang/pkg/services"
	"github.com/gin-gonic/gin"
)

type QueryRequest struct {
	Content string `json:"content" binding:"required"`
}

func Query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	trimmedContent := strings.TrimSpace(req.Content)
	if len(trimmedContent) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query content must be at least 3 characters long"})
		return
	}

	// Каналы для результатов и ошибок
	embeddingChan := make(chan []float32, 1)
	errChan := make(chan error, 2)

	// Асинхронно получаем эмбеддинг
	go func() {
		embedding, err := services.GetEmbedding(trimmedContent)
		if err != nil {
			errChan <- err
			return
		}
		embeddingChan <- embedding
	}()

	// Ждем эмбеддинг
	var queryEmbedding []float32
	select {
	case embedding := <-embeddingChan:
		queryEmbedding = embedding
	case err := <-errChan:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get query embedding: " + err.Error()})
		return
	}

	// Асинхронно ищем документы
	contextChan := make(chan string, 1)
	go func() {
		context, err := services.SearchDocuments(queryEmbedding)
		if err != nil {
			errChan <- err
			return
		}
		contextChan <- context
	}()

	// Ждем контекст
	var context string
	select {
	case ctx := <-contextChan:
		context = ctx
	case err := <-errChan:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search documents: " + err.Error()})
		return
	}

	// Генерируем ответ с помощью LLM
	response, err := services.GenerateResponse(trimmedContent, context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}
