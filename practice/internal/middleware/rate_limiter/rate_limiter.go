package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// TokenBucketLimiter implements a distributed token bucket rate limiter using Redis
type TokenBucketLimiter struct {
	redisClient *redis.Client
	prefix      string
	rate        float64 // Tokens per second
	capacity    int     // Maximum bucket size
}

// NewTokenBucketLimiter creates a new distributed token bucket rate limiter
func NewTokenBucketLimiter(
	redisClient *redis.Client,
	prefix string,
	rate float64,
	capacity int,
) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		redisClient: redisClient,
		prefix:      prefix,
		rate:        rate,
		capacity:    capacity,
	}
}

// Allow checks if a request should be allowed and consumes a token if available
// Returns: allowed, remaining tokens, time until next token, error
func (l *TokenBucketLimiter) Allow(
	ctx context.Context,
	key string,
) (bool, float64, time.Duration, error) {
	// Create Redis keys
	tokenKey := fmt.Sprintf("%s:%s:tokens", l.prefix, key)
	timestampKey := fmt.Sprintf("%s:%s:ts", l.prefix, key)

	// Execute token bucket algorithm in Lua script to ensure atomicity
	script := `
	local tokens_key = KEYS[1]
	local timestamp_key = KEYS[2]
	
	local rate = tonumber(ARGV[1])
	local capacity = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])
	
	-- Initialize if not exists
	local last_tokens = tonumber(redis.call("get", tokens_key))
	if last_tokens == nil then
		last_tokens = capacity
	end
	
	local last_refreshed = tonumber(redis.call("get", timestamp_key))
	if last_refreshed == nil then
		last_refreshed = now
	end
	
	-- Calculate the new token count
	local time_passed = math.max(0, now - last_refreshed)
	local new_tokens = last_tokens + (rate * time_passed)
	
	-- Cap tokens to the capacity
	new_tokens = math.min(capacity, new_tokens)
	
	-- If not enough tokens, return 0 (not allowed)
	local allowed = 0
	local new_tokens_after_request = new_tokens
	local wait_time = 0
	
	if new_tokens >= requested then
		-- Allow the request and consume tokens
		new_tokens_after_request = new_tokens - requested
		allowed = 1
	else
		-- Calculate time until enough tokens will be available
		wait_time = (requested - new_tokens) / rate
	end
	
	-- Update token count and timestamp if allowed
	if allowed == 1 then
		redis.call("set", tokens_key, new_tokens_after_request)
		redis.call("set", timestamp_key, now)
		-- Set expiration on keys to clean up
		local ttl = math.ceil(capacity / rate * 2)
		redis.call("expire", tokens_key, ttl)
		redis.call("expire", timestamp_key, ttl)
	end
	
	return {allowed, new_tokens_after_request, wait_time}
	`

	// Current time in seconds with millisecond precision
	now := float64(time.Now().UnixNano()) / 1e9

	// Run the Lua script in Redis
	result, err := l.redisClient.Eval(ctx, script, []string{tokenKey, timestampKey},
		l.rate, l.capacity, now, 1).Result()
	if err != nil {
		return false, 0, 0, err
	}

	// Parse the result
	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		return false, 0, 0, errors.New("invalid response from Redis")
	}

	// Extract values
	allowed, _ := strconv.ParseBool(fmt.Sprintf("%v", values[0]))
	remaining, _ := strconv.ParseFloat(fmt.Sprintf("%v", values[1]), 64)
	waitTime, _ := strconv.ParseFloat(fmt.Sprintf("%v", values[2]), 64)

	return allowed, remaining, time.Duration(waitTime * float64(time.Second)), nil
}

// Middleware creates an HTTP middleware for rate limiting
func (l *TokenBucketLimiter) Middleware(
	keyFunc func(*http.Request) string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			allowed, remaining, retry, err := l.Allow(r.Context(), key)
			if err != nil {
				// Log error but allow request to proceed in case of Redis failure
				// In production, you might want to implement a fallback strategy
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.capacity))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatFloat(remaining, 'f', 2, 64))

			if !allowed {
				retrySeconds := int(retry.Seconds() + 0.5) // Round to nearest second
				w.Header().Set("Retry-After", strconv.Itoa(retrySeconds))
				w.Header().
					Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(retry).Unix(), 10))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AllowN checks if a request should be allowed and consumes N tokens if available
func (l *TokenBucketLimiter) AllowN(
	ctx context.Context,
	key string,
	n int,
) (bool, float64, time.Duration, error) {
	// Implementation is similar to Allow but requests n tokens instead of 1
	// This is useful for API endpoints that have different costs

	tokenKey := fmt.Sprintf("%s:%s:tokens", l.prefix, key)
	timestampKey := fmt.Sprintf("%s:%s:ts", l.prefix, key)

	script := `
	-- Same script as above but with requested tokens from ARGV[4]
	local tokens_key = KEYS[1]
	local timestamp_key = KEYS[2]
	
	local rate = tonumber(ARGV[1])
	local capacity = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])
	
	local last_tokens = tonumber(redis.call("get", tokens_key))
	if last_tokens == nil then
		last_tokens = capacity
	end
	
	local last_refreshed = tonumber(redis.call("get", timestamp_key))
	if last_refreshed == nil then
		last_refreshed = now
	end
	
	local time_passed = math.max(0, now - last_refreshed)
	local new_tokens = last_tokens + (rate * time_passed)
	new_tokens = math.min(capacity, new_tokens)
	
	local allowed = 0
	local new_tokens_after_request = new_tokens
	local wait_time = 0
	
	if new_tokens >= requested then
		new_tokens_after_request = new_tokens - requested
		allowed = 1
	else
		wait_time = (requested - new_tokens) / rate
	end
	
	if allowed == 1 then
		redis.call("set", tokens_key, new_tokens_after_request)
		redis.call("set", timestamp_key, now)
		local ttl = math.ceil(capacity / rate * 2)
		redis.call("expire", tokens_key, ttl)
		redis.call("expire", timestamp_key, ttl)
	end
	
	return {allowed, new_tokens_after_request, wait_time}
	`

	now := float64(time.Now().UnixNano()) / 1e9

	result, err := l.redisClient.Eval(ctx, script, []string{tokenKey, timestampKey},
		l.rate, l.capacity, now, n).Result()
	if err != nil {
		return false, 0, 0, err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		return false, 0, 0, errors.New("invalid response from Redis")
	}

	allowed, _ := strconv.ParseBool(fmt.Sprintf("%v", values[0]))
	remaining, _ := strconv.ParseFloat(fmt.Sprintf("%v", values[1]), 64)
	waitTime, _ := strconv.ParseFloat(fmt.Sprintf("%v", values[2]), 64)

	return allowed, remaining, time.Duration(waitTime * float64(time.Second)), nil
}

// Apply rate_limiter as middleware
func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	// Create rate limiters for different tiers
	freeUserLimiter := ratelimit.NewRedisRateLimiter(rdb, "free_tier", 60, time.Minute)
	premiumUserLimiter := ratelimit.NewRedisRateLimiter(rdb, "premium_tier", 600, time.Minute)

	// Set up router (using gorilla/mux as an example)
	r := mux.NewRouter()

	// Apply rate limiters based on user tier
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := getUserFromRequest(r) // Your authentication logic

			var limiterMiddleware func(http.Handler) http.Handler

			if user.IsPremium {
				limiterMiddleware = premiumUserLimiter.Middleware(func(r *http.Request) string {
					return user.ID
				})
			} else {
				limiterMiddleware = freeUserLimiter.Middleware(func(r *http.Request) string {
					return user.ID
				})
			}

			limiterMiddleware(next).ServeHTTP(w, r)
		})
	})

	// Add your API routes
	r.HandleFunc("/api/resource", handleResource).Methods("GET")

	// Start server
	http.ListenAndServe(":8080", r)
}
