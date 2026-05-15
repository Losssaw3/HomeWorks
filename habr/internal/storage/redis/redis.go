package redishandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sso/internal/domain/models"
	"sso/internal/storage"

	"github.com/redis/go-redis/v9"
)

type CacheStorage struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCacheStorage(addr string, password string, dbNum int, ttl time.Duration) (*CacheStorage, error) {
	const op = "storage.redis.newCacheRedis"
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &CacheStorage{client: rdb, ttl: ttl}, nil
}

func (c *CacheStorage) GetArticleFromCacheById(ctx context.Context, key string) (models.Article, error) {
	const op = "storage.redis.GetArticleFromCacheById"
	val, err := c.client.Get(ctx, key).Bytes()
	var article models.Article
	if err == nil {
		err := json.Unmarshal(val, &article)
		if err != nil {
			return models.Article{}, fmt.Errorf("%s: %w", op, err)
		}
		return article, nil
	}
	if errors.Is(err, redis.Nil) {
		return models.Article{}, storage.ErrCacheNotFound
	} else {
		return models.Article{}, fmt.Errorf("%s: %w", op, err)
	}
}

func (c *CacheStorage) GetArticleFromCacheByAuthor(ctx context.Context, key string) ([]models.Article, error) {
	const op = "storage.redis.GetArticlesFromCacheByAuthor"
	val, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, storage.ErrCacheNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var articles []models.Article
	if err := json.Unmarshal(val, &articles); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return articles, nil
}

func (c *CacheStorage) AddArticleToChache(ctx context.Context, key string, in models.Article) error {
	const op = "storage.redis.AddArticlesToCache"
	data, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *CacheStorage) AddArticlesToCacheByAuthor(ctx context.Context, key string, articles []models.Article) error {
	const op = "storage.redis.AddArticlesToCacheByAuthor"
	data, err := json.Marshal(articles)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (c *CacheStorage) DeleteArticleFromCache(ctx context.Context, keys []string) error {
	const op = "storage.DeleteArticleFromCache"
	for _, key := range keys {
		err := c.client.Del(ctx, key).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CacheStorage) Close() error {
	const op = "storage.cache.Stop"

	if c.client == nil {
		return nil
	}
	if err := c.client.Close(); err != nil {
		return err
	}
	return nil
}
