package models

import "time"

// Fakultas mewakili struktur fakultas dalam database
type Fakultas struct {
	ID        string    `json:"id"`
	Nama      string    `json:"nama"`
	Kode      string    `json:"kode"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateFakultasRequest adalah request body untuk membuat fakultas baru
type CreateFakultasRequest struct {
	Nama string `json:"nama"`
	Kode string `json:"kode"`
}

// UpdateFakultasRequest adalah request body untuk update fakultas
type UpdateFakultasRequest struct {
	Nama string `json:"nama"`
	Kode string `json:"kode"`
}
