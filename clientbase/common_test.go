package clientbase

import (
	"fmt"
	"net/http"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	notFound := &APIError{StatusCode: http.StatusNotFound}
	serverError := &APIError{StatusCode: http.StatusInternalServerError}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"direct not found", notFound, true},
		{"wrapped not found", fmt.Errorf("operation failed: %w", notFound), true},
		{"non-404 error", serverError, false},
		{"wrapped non-404 error", fmt.Errorf("operation failed: %w", serverError), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
