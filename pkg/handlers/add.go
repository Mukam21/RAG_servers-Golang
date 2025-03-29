package handlers

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if len(req.Documents) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Documents list cannot be empty"})
		return
	}

	errChan := make(chan error, len(req.Documents))
	var wg sync.WaitGroup

	for _, doc := range req.Documents {
		trimmedText := strings.TrimSpace(doc.Text)
		if len(trimmedText) < 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Document text must be at least 5 characters long"})
			return
		}

		wg.Add(1)
		go func(text string) {
			defer wg.Done()
			embedding, err := services.GetEmbedding(text)
			if err != nil {
				errChan <- err
				return
			}
			if err := services.AddDocument(text, embedding); err != nil {
				errChan <- err
			}
		}(trimmedText)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process documents: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Documents added successfully"})
}
