package models

import (
	"time"
)

type Department struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	ParentID  *int      `gorm:"column:parent_id" json:"parent_id,omitempty"` // указатель для NULL
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Связи
	Parent    *Department  `gorm:"foreignKey:ParentID;references:ID" json:"-"` // обратная связь
	Children  []Department `gorm:"foreignKey:ParentID;references:ID" json:"children,omitempty"`
	Employees []Employee   `json:"employees" gorm:"foreignKey:DepartmentID;references:ID"`
}

func (Department) TableName() string {
	return "departments"
}

type Employee struct {
	ID           int        `json:"id" gorm:"primaryKey"`
	DepartmentID int        `json:"department_id" gorm:"not null;index"`
	FullName     string     `json:"full_name" gorm:"type:varchar(255);not null"`
	Position     string     `json:"position" gorm:"type:varchar(255);not null"`
	HiredAt      *time.Time `json:"hired_at"` // указатель для NULL
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Связь
	Department *Department `json:"department" gorm:"foreignKey:DepartmentID;references:ID"`
}

func (Employee) TableName() string {
	return "employees"
}
