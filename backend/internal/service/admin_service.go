package service

import (
	"errors"
	"net/mail"
	"strings"

	"employee-portal/backend/internal/domain"
	"employee-portal/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	repo *repository.ProfileRepository
}

type EmployeeDTO struct {
	ID       uint        `json:"id"`
	Email    string      `json:"email"`
	Role     string      `json:"role"`
	IsActive bool        `json:"isActive"`
	Profile  *ProfileDTO `json:"profile"`
}

type CreateEmployeeRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	MobilePhone  string `json:"mobilePhone"`
	ContactEmail string `json:"contactEmail"`
	Address      string `json:"address"`
	SkillIDs     []uint `json:"skillIds"`
}

func NewAdminService(repo *repository.ProfileRepository) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) ListEmployees() ([]EmployeeDTO, error) {
	users, err := s.repo.ListEmployees()
	if err != nil {
		return nil, err
	}
	out := make([]EmployeeDTO, 0, len(users))
	for _, user := range users {
		out = append(out, toEmployeeDTO(&user))
	}
	return out, nil
}

func (s *AdminService) CreateEmployee(req CreateEmployeeRequest) (*EmployeeDTO, error) {
	req.Email = strings.TrimSpace(req.Email)
	req.ContactEmail = strings.TrimSpace(req.ContactEmail)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.MobilePhone = strings.TrimSpace(req.MobilePhone)
	req.Address = strings.TrimSpace(req.Address)
	if req.Role == "" {
		req.Role = "employee"
	}
	if req.Role != "employee" && req.Role != "admin" {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Role must be employee or admin")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Login email format is invalid")
	}
	if req.ContactEmail == "" {
		req.ContactEmail = req.Email
	}
	if _, err := mail.ParseAddress(req.ContactEmail); err != nil {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Contact email format is invalid")
	}
	if len(req.Password) < 8 {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Password must be at least 8 characters")
	}
	if req.FirstName == "" || req.LastName == "" {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "First name and last name are required")
	}
	if len(req.Address) < 5 {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Address must be at least 5 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.CreateEmployee(repository.CreateEmployeeInput{
		Email: req.Email, PasswordHash: string(hash), Role: req.Role,
		FirstName: req.FirstName, LastName: req.LastName,
		MobilePhone: req.MobilePhone, ContactEmail: req.ContactEmail,
		Address: req.Address, SkillIDs: req.SkillIDs,
	})
	if err != nil {
		return nil, err
	}
	dto := toEmployeeDTO(user)
	return &dto, nil
}

func (s *AdminService) SetEmployeeActive(userID uint, active bool) (*EmployeeDTO, error) {
	user, err := s.repo.SetEmployeeActive(userID, active)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.NewAppError(domain.ErrNotFound, "NOT_FOUND", "Employee not found")
	}
	if err != nil {
		return nil, err
	}
	dto := toEmployeeDTO(user)
	return &dto, nil
}

func toEmployeeDTO(user *domain.User) EmployeeDTO {
	var profile *ProfileDTO
	if user.Profile != nil {
		profile = toProfileDTO(user.Profile)
	}
	return EmployeeDTO{
		ID: user.ID, Email: user.Email, Role: user.Role,
		IsActive: user.IsActive, Profile: profile,
	}
}
