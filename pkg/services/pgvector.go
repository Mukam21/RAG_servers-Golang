package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var pgConn *pgx.Conn

func init() {
	var err error
	// Подключаемся к PostgreSQL
	connStr := "postgres://postgres:raggolang@localhost:5438/postgres?sslmode=disable"
	pgConn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to PostgreSQL: %v", err))
	}

	_, err = pgConn.Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		panic(fmt.Sprintf("failed to create pgvector extension: %v", err))
	}

	_, err = pgConn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS documents (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			embedding VECTOR(768) -- Размерность вектора зависит от модели (например, 768 для Gemini)
		)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to create documents table: %v", err))
	}

	_, err = pgConn.Exec(context.Background(), `
		CREATE INDEX IF NOT EXISTS documents_embedding_idx ON documents USING hnsw (embedding vector_l2_ops)
	`)
	if err != nil {
		panic(fmt.Sprintf("failed to create index: %v", err))
	}
}

// AddDocument добавляет документ в PostgreSQL
func AddDocument(content string, embedding []float32) error {
	_, err := pgConn.Exec(context.Background(), `
		INSERT INTO documents (content, embedding)
		VALUES ($1, $2)
	`, content, embedding)
	return err
}

// SearchDocuments ищет наиболее релевантные документы
func SearchDocuments(queryEmbedding []float32) (string, error) {
	var content string
	err := pgConn.QueryRow(context.Background(), `
		SELECT content
		FROM documents
		ORDER BY embedding <-> $1
		LIMIT 1
	`, queryEmbedding).Scan(&content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no documents found")
		}
		return "", fmt.Errorf("failed to search documents: %v", err)
	}
	return content, nil
}

// CloseConnection закрывает соединение с PostgreSQL
func CloseConnection() {
	if pgConn != nil {
		pgConn.Close(context.Background())
	}
}
