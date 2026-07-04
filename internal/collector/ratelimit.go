package collector

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v88/github"
)

// RateLimitConfig configures rate limit handling behavior.
type RateLimitConfig struct {
	// MinRemaining is the minimum remaining requests before pausing.
	// When rate limit remaining falls below this, we wait until reset.
	MinRemaining int

	// MaxRetries is the maximum number of retries for rate-limited requests.
	MaxRetries int

	// BaseBackoff is the base duration for exponential backoff.
	BaseBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration

	// Logger for rate limit warnings (optional).
	Logger *slog.Logger
}

// DefaultRateLimitConfig returns sensible default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MinRemaining: 100,
		MaxRetries:   5,
		BaseBackoff:  1 * time.Second,
		MaxBackoff:   5 * time.Minute,
	}
}

// RateLimiter handles GitHub API rate limiting with exponential backoff.
type RateLimiter struct {
	config RateLimitConfig
}

// NewRateLimiter creates a new rate limiter with the given configuration.
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{config: config}
}

// CheckRateLimit checks the current rate limit status and waits if necessary.
// Returns an error if the context is cancelled while waiting.
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, client *github.Client) error {
	limits, _, err := client.RateLimit.Get(ctx)
	if err != nil {
		// If we can't get rate limits, proceed with caution
		rl.logWarn("failed to get rate limits, proceeding anyway", "error", err)
		return nil
	}

	core := limits.Core
	if core.Remaining < rl.config.MinRemaining {
		waitDuration := time.Until(core.Reset.Time)
		if waitDuration > 0 {
			rl.logWarn("rate limit low, waiting for reset",
				"remaining", core.Remaining,
				"limit", core.Limit,
				"reset", core.Reset.Time,
				"wait", waitDuration,
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
				return nil
			}
		}
	}

	return nil
}

// HandleRateLimitError handles a rate limit error response.
// It waits for the appropriate duration before retrying.
// Returns true if the request should be retried, false otherwise.
func (rl *RateLimiter) HandleRateLimitError(ctx context.Context, resp *github.Response, attempt int) (bool, error) {
	if attempt >= rl.config.MaxRetries {
		return false, nil
	}

	var waitDuration time.Duration

	// Check for Retry-After header (used for secondary rate limits)
	if resp != nil && resp.Header != nil {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
				waitDuration = time.Duration(seconds) * time.Second
			}
		}

		// Check for X-RateLimit-Reset header
		if waitDuration == 0 {
			if resetStr := resp.Header.Get("X-RateLimit-Reset"); resetStr != "" {
				if resetEpoch, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
					resetTime := time.Unix(resetEpoch, 0)
					waitDuration = time.Until(resetTime)
				}
			}
		}
	}

	// Use exponential backoff if no specific duration from headers
	if waitDuration <= 0 {
		waitDuration = rl.calculateBackoff(attempt)
	}

	// Cap at max backoff
	if waitDuration > rl.config.MaxBackoff {
		waitDuration = rl.config.MaxBackoff
	}

	rl.logWarn("rate limited, waiting before retry",
		"attempt", attempt+1,
		"maxRetries", rl.config.MaxRetries,
		"wait", waitDuration,
	)

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(waitDuration):
		return true, nil
	}
}

// calculateBackoff returns the backoff duration for the given attempt number.
// Uses exponential backoff with jitter.
func (rl *RateLimiter) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: base * 2^attempt
	backoff := float64(rl.config.BaseBackoff) * math.Pow(2, float64(attempt))

	// Add jitter (±25%) - not security-sensitive, math/rand is fine
	jitter := backoff * 0.25 * (rand.Float64()*2 - 1) //nolint:gosec
	backoff += jitter

	return time.Duration(backoff)
}

// IsRateLimitError checks if an error is a GitHub rate limit error.
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	// Check for GitHub rate limit error
	if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		_ = rateLimitErr
		return true
	}

	// Check for abuse rate limit error (secondary rate limit)
	if abuseErr, ok := err.(*github.AbuseRateLimitError); ok {
		_ = abuseErr
		return true
	}

	return false
}

// IsRateLimitResponse checks if a response indicates rate limiting.
func IsRateLimitResponse(resp *github.Response) bool {
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode == http.StatusForbidden
}

// GetRateLimitRemaining returns the remaining rate limit from a response.
func GetRateLimitRemaining(resp *github.Response) int {
	if resp == nil || resp.Header == nil {
		return -1
	}

	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if remaining == "" {
		return -1
	}

	n, err := strconv.Atoi(remaining)
	if err != nil {
		return -1
	}

	return n
}

// logWarn logs a warning message if a logger is configured.
func (rl *RateLimiter) logWarn(msg string, args ...any) {
	if rl.config.Logger != nil {
		rl.config.Logger.Warn(msg, args...)
	}
}
