package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"employee-portal/backend/internal/config"
	"employee-portal/backend/internal/domain"
	"employee-portal/backend/internal/repository"
	"gorm.io/gorm"
)

type ProfileService struct {
	cfg  config.Config
	repo *repository.ProfileRepository
}

type UserDTO struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type SkillDTO struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProfileDTO struct {
	ID           uint       `json:"id"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	AvatarURL    string     `json:"avatarUrl"`
	MobilePhone  string     `json:"mobilePhone"`
	ContactEmail string     `json:"contactEmail"`
	Address      string     `json:"address"`
	Skills       []SkillDTO `json:"skills"`
}

type ChangeLogDTO struct {
	ID         uint   `json:"id"`
	ChangeType string `json:"changeType"`
	FieldName  string `json:"fieldName"`
	OldValue   string `json:"oldValue"`
	NewValue   string `json:"newValue"`
	CreatedAt  string `json:"createdAt"`
}

func NewProfileService(cfg config.Config, repo *repository.ProfileRepository) *ProfileService {
	return &ProfileService{cfg: cfg, repo: repo}
}

func (s *ProfileService) GetMe(userID uint) (*ProfileDTO, error) {
	profile, err := s.repo.FindProfileByUserID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.NewAppError(domain.ErrNotFound, "NOT_FOUND", "Profile not found")
	}
	if err != nil {
		return nil, err
	}
	return toProfileDTO(profile), nil
}

func (s *ProfileService) ListSkills() ([]SkillDTO, error) {
	skills, err := s.repo.ListActiveSkills()
	if err != nil {
		return nil, err
	}
	out := make([]SkillDTO, 0, len(skills))
	for _, skill := range skills {
		out = append(out, SkillDTO{ID: skill.ID, Code: skill.Code, Name: skill.Name})
	}
	return out, nil
}

func (s *ProfileService) UpdateContact(userID uint, mobile, email, address string) (*ProfileDTO, error) {
	mobile = strings.TrimSpace(mobile)
	email = strings.TrimSpace(email)
	address = strings.TrimSpace(address)
	if !regexp.MustCompile(`^[0-9+\-\s]{8,20}$`).MatchString(mobile) {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Mobile phone format is invalid")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Contact email format is invalid")
	}
	if len(address) < 5 {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Address must be at least 5 characters")
	}
	profile, _, err := s.repo.UpdateContact(userID, mobile, email, address)
	if err != nil {
		return nil, err
	}
	return s.GetMe(profile.UserID)
}

func (s *ProfileService) UpdateSkills(userID uint, skillIDs []uint) (*ProfileDTO, error) {
	profile, _, err := s.repo.UpdateSkills(userID, skillIDs)
	if err != nil {
		return nil, err
	}
	return toProfileDTO(profile), nil
}

func (s *ProfileService) ValidateAvatar(file *multipart.FileHeader) error {
	if file == nil {
		return domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Avatar file is required")
	}
	if file.Size > int64(s.cfg.UploadMaxMB)*1024*1024 {
		return domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", fmt.Sprintf("Avatar must be less than %d MB", s.cfg.UploadMaxMB))
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Avatar must be JPG, PNG, or WEBP")
	}
	return nil
}

func (s *ProfileService) UpdateAvatar(userID uint, avatarURL string) (*ProfileDTO, error) {
	profile, _, err := s.repo.UpdateAvatar(userID, avatarURL)
	if err != nil {
		return nil, err
	}
	return s.GetMe(profile.UserID)
}

func (s *ProfileService) ListHistory(userID uint, days int) ([]ChangeLogDTO, error) {
	if days <= 0 || days > s.cfg.HistoryRetentionDays {
		days = s.cfg.HistoryRetentionDays
	}
	since := time.Now().AddDate(0, 0, -days)
	logs, err := s.repo.ListHistory(userID, since)
	if err != nil {
		return nil, err
	}
	out := make([]ChangeLogDTO, 0, len(logs))
	for _, log := range logs {
		out = append(out, ChangeLogDTO{
			ID: log.ID, ChangeType: log.ChangeType, FieldName: log.FieldName,
			OldValue: log.OldValue, NewValue: log.NewValue,
			CreatedAt: log.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

func (s *ProfileService) ResetDemo(userID uint) (*ProfileDTO, error) {
	if s.cfg.AppEnv == "production" {
		return nil, domain.NewAppError(domain.ErrForbidden, "FORBIDDEN", "Demo reset is disabled in production")
	}
	profile, err := s.repo.ResetDemoUser(userID)
	if err != nil {
		return nil, err
	}
	return toProfileDTO(profile), nil
}

func toProfileDTO(profile *domain.EmployeeProfile) *ProfileDTO {
	skills := make([]SkillDTO, 0, len(profile.Skills))
	for _, skill := range profile.Skills {
		skills = append(skills, SkillDTO{ID: skill.ID, Code: skill.Code, Name: skill.Name})
	}
	return &ProfileDTO{
		ID: profile.ID, FirstName: profile.FirstName, LastName: profile.LastName,
		AvatarURL: profile.AvatarURL, MobilePhone: profile.MobilePhone,
		ContactEmail: profile.ContactEmail, Address: profile.Address, Skills: skills,
	}
}
