package cache

import (
	"context"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sangtandoan/social/internal/config"
)

var (
	LockValue      = "lock"
	ExpirationTime = time.Minute * 10
)

func NewRedisClient(cfg *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.DB,
		Password: cfg.Password,
	})
}

type CacheService struct {
	client *redis.Client
}

func NewCacheService(client *redis.Client) *CacheService {
	return &CacheService{client}
}

func (s *CacheService) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *CacheService) Set(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) error {
	random := rand.New(rand.NewSource(time.Now().Unix()))

	expiration = expiration + time.Duration(random.Intn(100)*int(time.Second))
	return s.client.SetEx(ctx, key, value, expiration).Err()
}

func (s *CacheService) SetNX(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) error {
	random := rand.New(rand.NewSource(time.Now().Unix()))
	expiration = expiration + time.Duration(random.Intn(100)*int(time.Second))
	// Prevent cache avalanche problem
	return s.client.SetNX(ctx, key, value, expiration).Err()
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}
