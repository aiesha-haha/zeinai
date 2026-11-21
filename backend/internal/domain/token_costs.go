package domain

// AI Token costs for different operations
const (
	// KMJ Generation cost
	TokenCostKMJGeneration = 5

	// Future AI operations (placeholder for future features)
	TokenCostSlideGeneration = 5
	TokenCostVideoGeneration = 10
	TokenCostImageGeneration = 2
	TokenCostTestGeneration  = 4
	TokenCostArticleGeneration = 3
)

// Action types for token usage logging
const (
	ActionTypeKMJGeneration   = "kmj_generation"
	ActionTypeSlideGeneration = "slide_generation"
	ActionTypeVideoGeneration = "video_generation"
	ActionTypeImageGeneration = "image_generation"
	ActionTypeTestGeneration  = "test_generation"
	ActionTypeArticleGeneration = "article_generation"
)
