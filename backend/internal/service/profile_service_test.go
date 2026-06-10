package service

import (
	"errors"
	"testing"

	"employee-portal/backend/internal/config"
	"employee-portal/backend/internal/domain"
)

func TestValidateAvatarRequiresFile(t *testing.T) {
	service := NewProfileService(config.Config{UploadMaxMB: 3}, nil)
	err := service.ValidateAvatar(nil)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestResetDemoDisabledInProduction(t *testing.T) {
	service := NewProfileService(config.Config{AppEnv: "production"}, nil)
	_, err := service.ResetDemo(1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}
