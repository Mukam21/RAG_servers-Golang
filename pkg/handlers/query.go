package handlers

import (
	"net/http"
	"strings"

	"github.com/Mukam21/RAG_server-Golang/pkg/services"
	"github.com/gin-gonic/gin"
)

type QueryRequest struct {
	Query string `json:"query" binding:"required"`
}

func Query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	trimmedQuery := strings.TrimSpace(req.Query)
	if len(trimmedQuery) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query must be at least 3 characters long"})
		return
	}

	embedding, err := services.GetEmbedding(trimmedQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get query embedding: " + err.Error()})
		return
	}

	context, err := services.SearchDocuments(embedding)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search documents: " + err.Error()})
		return
	}

	response, err := services.GenerateResponse(trimmedQuery, context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": response})
}
