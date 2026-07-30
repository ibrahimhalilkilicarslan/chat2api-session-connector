package browser

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNormalizeStoredToken(t *testing.T) {
	tests := map[string]struct {
		stored      tokenProbe
		expected    tokenProbe
		expectError bool
	}{
		"current DeepSeek JSON envelope": {
			stored: tokenProbe{
				State: "stored",
				Token: `{"value":"current-session-token","__version":"1"}`,
			},
			expected: tokenProbe{State: "stored", Token: "current-session-token"},
		},
		"legacy plain token": {
			stored:   tokenProbe{State: "stored", Token: " legacy-session-token "},
			expected: tokenProbe{State: "stored", Token: "legacy-session-token"},
		},
		"empty storage": {
			stored:   tokenProbe{State: "empty"},
			expected: tokenProbe{State: "empty"},
		},
		"malformed JSON envelope": {
			stored:   tokenProbe{State: "stored", Token: `{"value":`},
			expected: tokenProbe{State: "rejected"},
		},
		"JSON envelope without value": {
			stored:   tokenProbe{State: "stored", Token: `{"__version":"1"}`},
			expected: tokenProbe{State: "rejected"},
		},
		"sentinel value": {
			stored:   tokenProbe{State: "stored", Token: "undefined"},
			expected: tokenProbe{State: "rejected"},
		},
		"oversized storage value": {
			stored:      tokenProbe{State: "stored", Token: strings.Repeat("x", 16_385)},
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := normalizeStoredToken(test.stored)
			if test.expectError {
				if err == nil {
					t.Fatal("normalizeStoredToken() unexpectedly succeeded")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeStoredToken() error = %v", err)
			}
			if actual != test.expected {
				t.Fatalf("normalizeStoredToken() = %#v, want %#v", actual, test.expected)
			}
		})
	}
}

func TestWaitForTokenSurvivesHumanVerificationPlaceholder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reads := 0
	placeholderValidations := 0
	token, err := waitForToken(
		ctx,
		context.Background(),
		time.Millisecond,
		50*time.Millisecond,
		func(context.Context) (tokenProbe, error) {
			reads++
			if reads <= 5 {
				return tokenProbe{State: "stored", Token: "human-verification-placeholder"}, nil
			}
			return tokenProbe{State: "stored", Token: "signed-in-session-token"}, nil
		},
		func(_ context.Context, candidate string) (tokenProbe, error) {
			if candidate == "human-verification-placeholder" {
				placeholderValidations++
				return tokenProbe{State: "rejected"}, nil
			}
			return tokenProbe{State: "ready", Token: candidate}, nil
		},
	)
	if err != nil {
		t.Fatalf("waitForToken() error = %v", err)
	}
	if token != "signed-in-session-token" {
		t.Fatalf("waitForToken() token = %q", token)
	}
	if placeholderValidations != 1 {
		t.Fatalf("placeholder validations = %d, want 1", placeholderValidations)
	}
}

func TestWaitForTokenSurvivesRepeatedUnavailableTokens(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reads := 0
	token, err := waitForToken(
		ctx,
		context.Background(),
		time.Millisecond,
		time.Millisecond,
		func(context.Context) (tokenProbe, error) {
			reads++
			if reads <= 15 {
				return tokenProbe{State: "stored", Token: "pending-token-" + strconv.Itoa(reads)}, nil
			}
			return tokenProbe{State: "stored", Token: "accepted-session-token"}, nil
		},
		func(_ context.Context, candidate string) (tokenProbe, error) {
			if candidate == "accepted-session-token" {
				return tokenProbe{State: "ready", Token: candidate}, nil
			}
			return tokenProbe{State: "unavailable"}, nil
		},
	)
	if err != nil {
		t.Fatalf("waitForToken() error = %v", err)
	}
	if token != "accepted-session-token" {
		t.Fatalf("waitForToken() token = %q", token)
	}
}
