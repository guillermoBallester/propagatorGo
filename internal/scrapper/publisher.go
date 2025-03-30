package scraper

import (
	"context"
	"fmt"
	"log"
	"propagatorGo/internal/queue"
)

const ArticleQueue = "article"

// ArticlePublisher handles publishing scraped articles to a message queue
type ArticlePublisher struct {
	redis *queue.RedisClient
}

// NewArticlePublisher creates a new article publisher
func NewArticlePublisher(redis *queue.RedisClient) *ArticlePublisher {
	return &ArticlePublisher{
		redis: redis,
	}
}

// PublishArticle publishes a single article to the queue
func (p *ArticlePublisher) PublishArticle(ctx context.Context, article ArticleData) error {
	err := p.redis.PublishMessage(ctx, ArticleQueue, article)
	if err != nil {
		return fmt.Errorf("failed to publish article: %w", err)
	}
	return nil
}

// PublishArticles publishes multiple articles to the queue
func (p *ArticlePublisher) PublishArticles(ctx context.Context, articles []ArticleData) error {
	for _, article := range articles {
		if err := p.PublishArticle(ctx, article); err != nil {
			log.Printf("Error publishing article '%s': %v", article.Title, err)
		}
	}
	return nil
}
