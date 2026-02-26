package models

import "time"

// StudentRegistration mewakili data pendaftaran mahasiswa yang menunggu verifikasi
type StudentRegistration struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Hidden from JSON
	KtmPath   string    `json:"ktm_path"`
	Status    string    `json:"status"` // "pending", "approved", "rejected"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
