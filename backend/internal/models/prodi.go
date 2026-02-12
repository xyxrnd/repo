package models

import "time"

// Prodi mewakili struktur program studi dalam database
type Prodi struct {
	ID           string    `json:"id"`
	Nama         string    `json:"nama"`
	Kode         string    `json:"kode"`
	FakultasID   string    `json:"fakultas_id"`
	FakultasNama string    `json:"fakultas_nama,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateProdiRequest adalah request body untuk membuat prodi baru
type CreateProdiRequest struct {
	Nama       string `json:"nama"`
	Kode       string `json:"kode"`
	FakultasID string `json:"fakultas_id"`
}

// UpdateProdiRequest adalah request body untuk update prodi
type UpdateProdiRequest struct {
	Nama       string `json:"nama"`
	Kode       string `json:"kode"`
	FakultasID string `json:"fakultas_id"`
}
