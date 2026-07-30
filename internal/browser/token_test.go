package browser

import (
	"strings"
	"testing"
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
