package domain

import (
	"github.com/google/uuid"
)

type CreateArticleRequest struct {
	Subject     string `json:"subject" validate:"required,max=255"`
	Grade       int    `json:"grade" validate:"required,min=1,max=12"`
	Topic       string `json:"topic" validate:"required,max=500"`
	Language    string `json:"language" validate:"required,oneof=kaz rus eng"`
	Description string `json:"description" validate:"max=1000"`
	Style       string `json:"style" validate:"required"`
	Complexity  string `json:"complexity" validate:"required"`
}

type ArticleResponse struct {
	ID          uuid.UUID     `json:"id"`
	Subject     string        `json:"subject"`
	Grade       int           `json:"grade"`
	Topic       string        `json:"topic"`
	Language    string        `json:"language"`
	Description string        `json:"description"`
	Style       string        `json:"style"`
	Complexity  string        `json:"complexity"`
	Title       *string       `json:"title,omitempty"`
	Content     *string       `json:"content,omitempty"`
	Status      ArticleStatus `json:"status"`
	CreatedAt   string        `json:"created_at"`
}
