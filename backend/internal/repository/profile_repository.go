package repository

import (
	"time"

	"employee-portal/backend/internal/domain"
	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) FindUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Preload("Profile").Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ProfileRepository) FindProfileByUserID(userID uint) (*domain.EmployeeProfile, error) {
	var profile domain.EmployeeProfile
	err := r.db.Preload("Skills", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepository) ListActiveSkills() ([]domain.Skill, error) {
	var skills []domain.Skill
	err := r.db.Where("is_active = ?", true).Order("display_order ASC").Find(&skills).Error
	return skills, err
}

func (r *ProfileRepository) ListEmployees() ([]domain.User, error) {
	var users []domain.User
	err := r.db.Preload("Profile.Skills", func(db *gorm.DB) *gorm.DB {
		return db.Order("display_order ASC")
	}).Order("created_at DESC").Find(&users).Error
	return users, err
}

type CreateEmployeeInput struct {
	Email        string
	PasswordHash string
	Role         string
	FirstName    string
	LastName     string
	MobilePhone  string
	ContactEmail string
	Address      string
	SkillIDs     []uint
}

func (r *ProfileRepository) CreateEmployee(input CreateEmployeeInput) (*domain.User, error) {
	var created domain.User
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&domain.User{}).Where("email = ?", input.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return domain.NewAppError(domain.ErrConflict, "CONFLICT", "Employee email already exists")
		}

		var skills []domain.Skill
		if len(input.SkillIDs) > 0 {
			if err := tx.Where("id IN ? AND is_active = ?", input.SkillIDs, true).Order("display_order ASC").Find(&skills).Error; err != nil {
				return err
			}
			if len(skills) != len(input.SkillIDs) {
				return domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "One or more skills are invalid")
			}
		}

		user := domain.User{
			Email:        input.Email,
			PasswordHash: input.PasswordHash,
			Role:         input.Role,
			IsActive:     true,
			Profile: &domain.EmployeeProfile{
				FirstName:    input.FirstName,
				LastName:     input.LastName,
				MobilePhone:  input.MobilePhone,
				ContactEmail: input.ContactEmail,
				Address:      input.Address,
				Skills:       skills,
			},
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Preload("Profile.Skills", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).First(&created, user.ID).Error
	})
	return &created, err
}

func (r *ProfileRepository) SetEmployeeActive(userID uint, active bool) (*domain.User, error) {
	var user domain.User
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		user.IsActive = active
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		return tx.Preload("Profile.Skills", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).First(&user, userID).Error
	})
	return &user, err
}

func (r *ProfileRepository) UpdateContact(userID uint, mobile, email, address string) (*domain.EmployeeProfile, []domain.ProfileChangeLog, error) {
	var updated domain.EmployeeProfile
	var logs []domain.ProfileChangeLog
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var profile domain.EmployeeProfile
		if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			return err
		}
		logs = appendChanged(logs, userID, profile.ID, "contact", "mobile_phone", profile.MobilePhone, mobile)
		logs = appendChanged(logs, userID, profile.ID, "contact", "contact_email", profile.ContactEmail, email)
		logs = appendChanged(logs, userID, profile.ID, "contact", "address", profile.Address, address)

		profile.MobilePhone = mobile
		profile.ContactEmail = email
		profile.Address = address
		if err := tx.Save(&profile).Error; err != nil {
			return err
		}
		if len(logs) > 0 {
			if err := tx.Create(&logs).Error; err != nil {
				return err
			}
		}
		updated = profile
		return nil
	})
	return &updated, logs, err
}

func (r *ProfileRepository) UpdateAvatar(userID uint, avatarURL string) (*domain.EmployeeProfile, *domain.ProfileChangeLog, error) {
	var updated domain.EmployeeProfile
	var log *domain.ProfileChangeLog
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var profile domain.EmployeeProfile
		if err := tx.Where("user_id = ?", userID).First(&profile).Error; err != nil {
			return err
		}
		if profile.AvatarURL != avatarURL {
			entry := domain.ProfileChangeLog{
				UserID: userID, EmployeeProfileID: profile.ID,
				ChangeType: "avatar", FieldName: "avatar_url",
				OldValue: profile.AvatarURL, NewValue: avatarURL,
			}
			log = &entry
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		profile.AvatarURL = avatarURL
		if err := tx.Save(&profile).Error; err != nil {
			return err
		}
		updated = profile
		return nil
	})
	return &updated, log, err
}

func (r *ProfileRepository) UpdateSkills(userID uint, skillIDs []uint) (*domain.EmployeeProfile, *domain.ProfileChangeLog, error) {
	var updated domain.EmployeeProfile
	var log *domain.ProfileChangeLog
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var profile domain.EmployeeProfile
		if err := tx.Preload("Skills", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).Where("user_id = ?", userID).First(&profile).Error; err != nil {
			return err
		}

		var skills []domain.Skill
		if len(skillIDs) > 0 {
			if err := tx.Where("id IN ? AND is_active = ?", skillIDs, true).Order("display_order ASC").Find(&skills).Error; err != nil {
				return err
			}
			if len(skills) != len(skillIDs) {
				return domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "One or more skills are invalid")
			}
		}

		oldValue := skillNames(profile.Skills)
		newValue := skillNames(skills)
		if oldValue != newValue {
			entry := domain.ProfileChangeLog{
				UserID: userID, EmployeeProfileID: profile.ID,
				ChangeType: "skills", FieldName: "skills",
				OldValue: oldValue, NewValue: newValue,
			}
			log = &entry
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&profile).Association("Skills").Replace(skills); err != nil {
			return err
		}
		return tx.Preload("Skills", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).First(&updated, profile.ID).Error
	})
	return &updated, log, err
}

func (r *ProfileRepository) ListHistory(userID uint, since time.Time) ([]domain.ProfileChangeLog, error) {
	var logs []domain.ProfileChangeLog
	err := r.db.Where("user_id = ? AND created_at >= ?", userID, since).Order("created_at DESC").Find(&logs).Error
	return logs, err
}

func (r *ProfileRepository) ResetDemoUser(userID uint) (*domain.EmployeeProfile, error) {
	var updated domain.EmployeeProfile
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var profile domain.EmployeeProfile
		if err := tx.Preload("Skills").Where("user_id = ?", userID).First(&profile).Error; err != nil {
			return err
		}

		var defaultSkills []domain.Skill
		if err := tx.Where("code IN ?", []string{"go", "react"}).Order("display_order ASC").Find(&defaultSkills).Error; err != nil {
			return err
		}

		profile.AvatarURL = ""
		profile.MobilePhone = "0812345678"
		profile.ContactEmail = "narin.dev@example.com"
		profile.Address = "Bangkok, Thailand"
		if err := tx.Save(&profile).Error; err != nil {
			return err
		}
		if err := tx.Model(&profile).Association("Skills").Replace(defaultSkills); err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&domain.ProfileChangeLog{}).Error; err != nil {
			return err
		}
		return tx.Preload("Skills", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).First(&updated, profile.ID).Error
	})
	return &updated, err
}

func appendChanged(logs []domain.ProfileChangeLog, userID, profileID uint, changeType, field, oldValue, newValue string) []domain.ProfileChangeLog {
	if oldValue == newValue {
		return logs
	}
	return append(logs, domain.ProfileChangeLog{
		UserID: userID, EmployeeProfileID: profileID,
		ChangeType: changeType, FieldName: field,
		OldValue: oldValue, NewValue: newValue,
	})
}

func skillNames(skills []domain.Skill) string {
	names := ""
	for i, skill := range skills {
		if i > 0 {
			names += ", "
		}
		names += skill.Name
	}
	return names
}
