package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Citation struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	PublicationDate string `json:"publication_date"`
	URL             string `json:"url"`
	Publisher       string `json:"publisher"`
	Version         string `json:"version"`
}

type CitationCache struct {
	redis *RedisClient
}

func NewCitationCache(r *RedisClient) *CitationCache {
	return &CitationCache{redis: r}
}

func (c *CitationCache) ListAll(ctx context.Context) ([]Citation, error) {
	var keys []string
	var cursor uint64

	for {
		batch, next, err := c.redis.client.Scan(ctx, cursor, "cite:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("scanning citation keys: %w", err)
		}
		keys = append(keys, batch...)
		cursor = next
		if cursor == 0 {
			break
		}
	}

	citations := make([]Citation, 0, len(keys))
	for _, key := range keys {
		fields, err := c.redis.client.HGetAll(ctx, key).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			return nil, fmt.Errorf("reading citation hash %s: %w", key, err)
		}

		citation := Citation{
			Title:           fields["title"],
			Author:          fields["author"],
			PublicationDate: fields["publication_date"],
			URL:             fields["url"],
			Publisher:       fields["publisher"],
			Version:         fields["version"],
		}
		citations = append(citations, citation)
	}

	return citations, nil
}

func (c *CitationCache) Get(ctx context.Context, key string) (*Citation, error) {
	fields, err := c.redis.client.HGetAll(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading citation hash %s: %w", key, err)
	}
	if len(fields) == 0 {
		return nil, nil
	}
	return &Citation{
		Title:           fields["title"],
		Author:          fields["author"],
		PublicationDate: fields["publication_date"],
		URL:             fields["url"],
		Publisher:       fields["publisher"],
		Version:         fields["version"],
	}, nil
}

func (c *CitationCache) Set(ctx context.Context, key string, citation Citation) error {
	err := c.redis.client.HSet(ctx, key,
		"title", citation.Title,
		"author", citation.Author,
		"publication_date", citation.PublicationDate,
		"url", citation.URL,
		"publisher", citation.Publisher,
		"version", citation.Version,
	).Err()
	if err != nil {
		return fmt.Errorf("storing citation hash %s: %w", key, err)
	}
	if err := c.redis.client.Expire(ctx, key, time.Hour).Err(); err != nil {
		return fmt.Errorf("setting TTL on citation hash %s: %w", key, err)
	}
	return nil
}
