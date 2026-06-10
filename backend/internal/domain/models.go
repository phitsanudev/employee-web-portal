package domain

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;size:160;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Role         string `gorm:"size:40;not null;default:employee"`
	IsActive     bool   `gorm:"not null;default:true"`
	Profile      *EmployeeProfile
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type EmployeeProfile struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint `gorm:"uniqueIndex;not null"`
	User         *User
	FirstName    string  `gorm:"size:120;not null"`
	LastName     string  `gorm:"size:120;not null"`
	AvatarURL    string  `gorm:"type:text"`
	MobilePhone  string  `gorm:"size:40"`
	ContactEmail string  `gorm:"size:160"`
	Address      string  `gorm:"type:text"`
	Skills       []Skill `gorm:"many2many:employee_profile_skills;"`
	ChangeLogs   []ProfileChangeLog
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Skill struct {
	ID           uint   `gorm:"primaryKey"`
	Code         string `gorm:"uniqueIndex;size:80;not null" json:"code"`
	Name         string `gorm:"size:120;not null" json:"name"`
	DisplayOrder int    `gorm:"not null;default:0" json:"display_order"`
	IsActive     bool   `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type EmployeeProfileSkill struct {
	EmployeeProfileID uint `gorm:"primaryKey"`
	SkillID           uint `gorm:"primaryKey"`
	CreatedAt         time.Time
}

type ProfileChangeLog struct {
	ID                uint      `gorm:"primaryKey"`
	UserID            uint      `gorm:"index;not null"`
	EmployeeProfileID uint      `gorm:"index;not null"`
	ChangeType        string    `gorm:"size:60;not null"`
	FieldName         string    `gorm:"size:80;not null"`
	OldValue          string    `gorm:"type:text"`
	NewValue          string    `gorm:"type:text"`
	CreatedAt         time.Time `gorm:"index"`
}
