package database

import (
	"encoding/json"
	"log/slog"
	"os"

	"employee-portal/backend/internal/config"
	"employee-portal/backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.SetupJoinTable(&domain.EmployeeProfile{}, "Skills", &domain.EmployeeProfileSkill{}); err != nil {
		return err
	}
	return db.AutoMigrate(
		&domain.User{},
		&domain.EmployeeProfile{},
		&domain.Skill{},
		&domain.EmployeeProfileSkill{},
		&domain.ProfileChangeLog{},
	)
}

func Seed(db *gorm.DB, logger *slog.Logger, cfg config.Config) error {
	skills, err := loadSkills(cfg.SkillsConfigPath)
	if err != nil {
		return err
	}
	for _, skill := range skills {
		if err := db.Where("code = ?", skill.Code).Assign(skill).FirstOrCreate(&skill).Error; err != nil {
			return err
		}
	}

	if err := seedUser(db, logger, "admin@employee.dev", "password123", "admin", "Admin", "Manager"); err != nil {
		return err
	}
	return seedUser(db, logger, "demo@employee.dev", "password123", "employee", "Narin", "Developer")
}

func loadSkills(path string) ([]domain.Skill, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var skills []domain.Skill
	if err := json.NewDecoder(file).Decode(&skills); err != nil {
		return nil, err
	}
	return skills, nil
}

func seedUser(db *gorm.DB, logger *slog.Logger, email, password, role, firstName, lastName string) error {
	var count int64
	if err := db.Model(&domain.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
		Profile: &domain.EmployeeProfile{
			FirstName:    firstName,
			LastName:     lastName,
			MobilePhone:  "0812345678",
			ContactEmail: email,
			Address:      "Bangkok, Thailand",
		},
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	logger.Info("seeded account", "email", email, "role", role, "password", password)
	return nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
