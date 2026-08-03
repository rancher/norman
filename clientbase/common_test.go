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

func TestIsUnauthorized(t *testing.T) {
	unauthorized := &APIError{StatusCode: http.StatusUnauthorized}
	serverError := &APIError{StatusCode: http.StatusInternalServerError}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"direct unauthorized", unauthorized, true},
		{"wrapped unauthorized", fmt.Errorf("operation failed: %w", unauthorized), true},
		{"non-401 error", serverError, false},
		{"wrapped non-401 error", fmt.Errorf("operation failed: %w", serverError), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.want {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsForbidden(t *testing.T) {
	forbidden := &APIError{StatusCode: http.StatusForbidden}
	serverError := &APIError{StatusCode: http.StatusInternalServerError}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"direct forbidden", forbidden, true},
		{"wrapped forbidden", fmt.Errorf("operation failed: %w", forbidden), true},
		{"non-403 error", serverError, false},
		{"wrapped non-403 error", fmt.Errorf("operation failed: %w", serverError), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbidden(tt.err); got != tt.want {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.want)
			}
		})
	}
}
