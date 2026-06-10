package response

import (
	"net/http"
	"testing"

	"employee-portal/backend/internal/domain"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{
			name:   "validation",
			err:    domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Invalid input"),
			status: http.StatusBadRequest,
			code:   "VALIDATION_ERROR",
		},
		{
			name:   "unauthorized",
			err:    domain.NewAppError(domain.ErrUnauthorized, "UNAUTHORIZED", "Invalid token"),
			status: http.StatusUnauthorized,
			code:   "UNAUTHORIZED",
		},
		{
			name:   "internal fallback",
			err:    domain.ErrInternal,
			status: http.StatusInternalServerError,
			code:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, _ := mapError(tt.err)
			if status != tt.status || code != tt.code {
				t.Fatalf("expected %d/%s, got %d/%s", tt.status, tt.code, status, code)
			}
		})
	}
}
