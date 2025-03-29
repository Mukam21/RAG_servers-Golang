package services

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

var pgConn *pgx.Conn

func init() {
	var err error
	password := os.Getenv("PG_PASSWORD")
	if password == "" {
		panic("PG_PASSWORD environment variable not set")
	}
	connStr := fmt.Sprintf("postgres://postgres:%s@localhost:5438/postgres?sslmode=disable", password)
	pgConn, err = pgx.Connect(context.Background(), connStr)
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
			embedding VECTOR(768)
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

func AddDocument(content string, embedding []float32) error {
	_, err := pgConn.Exec(context.Background(), `
		INSERT INTO documents (content, embedding)
		VALUES ($1, $2)
	`, content, embedding)
	return err
}

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

func CloseConnection() {
	if pgConn != nil {
		pgConn.Close(context.Background())
	}
}
