package handlers

import (
	"io"
	"net/http"

	"github.com/Mukam21/RAG_server-Golang/pkg/services"
	"github.com/gin-gonic/gin"
)

func UploadDocumentGin(c *gin.Context) {
	file, _, err := c.Request.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read content"})
		return
	}

	embedding, err := services.GetEmbedding(string(content))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate embedding: " + err.Error()})
		return
	}

	if err := services.AddDocument(string(content), embedding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save document: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document uploaded successfully"})
}
