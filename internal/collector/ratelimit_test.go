package collector

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-github/v88/github"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	if config.MinRemaining != 100 {
		t.Errorf("MinRemaining = %d, want 100", config.MinRemaining)
	}
	if config.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", config.MaxRetries)
	}
	if config.BaseBackoff != time.Second {
		t.Errorf("BaseBackoff = %v, want %v", config.BaseBackoff, time.Second)
	}
	if config.MaxBackoff != 5*time.Minute {
		t.Errorf("MaxBackoff = %v, want %v", config.MaxBackoff, 5*time.Minute)
	}
}

func TestNewRateLimiter(t *testing.T) {
	config := DefaultRateLimitConfig()
	rl := NewRateLimiter(config)

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
	if rl.config.MinRemaining != config.MinRemaining {
		t.Errorf("config.MinRemaining = %d, want %d", rl.config.MinRemaining, config.MinRemaining)
	}
}

func TestRateLimiter_calculateBackoff(t *testing.T) {
	config := RateLimitConfig{
		BaseBackoff: time.Second,
		MaxBackoff:  time.Minute,
	}
	rl := NewRateLimiter(config)

	tests := []struct {
		attempt     int
		wantMinBase time.Duration
		wantMaxBase time.Duration
	}{
		{0, 750 * time.Millisecond, 1250 * time.Millisecond},  // 1s ± 25%
		{1, 1500 * time.Millisecond, 2500 * time.Millisecond}, // 2s ± 25%
		{2, 3 * time.Second, 5 * time.Second},                 // 4s ± 25%
		{3, 6 * time.Second, 10 * time.Second},                // 8s ± 25%
	}

	for _, tt := range tests {
		t.Run("attempt_"+strconv.Itoa(tt.attempt), func(t *testing.T) {
			// Run multiple times to account for jitter
			for i := 0; i < 10; i++ {
				got := rl.calculateBackoff(tt.attempt)
				if got < tt.wantMinBase || got > tt.wantMaxBase {
					t.Errorf("calculateBackoff(%d) = %v, want between %v and %v",
						tt.attempt, got, tt.wantMinBase, tt.wantMaxBase)
				}
			}
		})
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"rate limit error", &github.RateLimitError{}, true},
		{"abuse rate limit error", &github.AbuseRateLimitError{}, true},
		{"generic error", &github.ErrorResponse{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRateLimitError(tt.err)
			if got != tt.want {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRateLimitResponse(t *testing.T) {
	tests := []struct {
		name string
		resp *github.Response
		want bool
	}{
		{
			name: "nil response",
			resp: nil,
			want: false,
		},
		{
			name: "rate limited (429)",
			resp: &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusTooManyRequests,
				},
			},
			want: true,
		},
		{
			name: "forbidden (403)",
			resp: &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusForbidden,
				},
			},
			want: true,
		},
		{
			name: "success (200)",
			resp: &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			},
			want: false,
		},
		{
			name: "not found (404)",
			resp: &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRateLimitResponse(tt.resp)
			if got != tt.want {
				t.Errorf("IsRateLimitResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRateLimitRemaining(t *testing.T) {
	tests := []struct {
		name string
		resp *github.Response
		want int
	}{
		{
			name: "nil response",
			resp: nil,
			want: -1,
		},
		{
			name: "valid remaining",
			resp: &github.Response{
				Response: &http.Response{
					Header: http.Header{
						"X-Ratelimit-Remaining": []string{"42"},
					},
				},
			},
			want: 42,
		},
		{
			name: "zero remaining",
			resp: &github.Response{
				Response: &http.Response{
					Header: http.Header{
						"X-Ratelimit-Remaining": []string{"0"},
					},
				},
			},
			want: 0,
		},
		{
			name: "missing header",
			resp: &github.Response{
				Response: &http.Response{
					Header: http.Header{},
				},
			},
			want: -1,
		},
		{
			name: "invalid value",
			resp: &github.Response{
				Response: &http.Response{
					Header: http.Header{
						"X-Ratelimit-Remaining": []string{"invalid"},
					},
				},
			},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRateLimitRemaining(tt.resp)
			if got != tt.want {
				t.Errorf("GetRateLimitRemaining() = %d, want %d", got, tt.want)
			}
		})
	}
}
