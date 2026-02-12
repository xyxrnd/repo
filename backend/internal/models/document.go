package models

import "time"

// Document mewakili struktur dokumen dalam database
type Document struct {
	ID              string         `json:"id"`
	Judul           string         `json:"judul"`
	Penulis         string         `json:"penulis"`
	JenisFile       string         `json:"jenis_file"`
	FakultasID      string         `json:"fakultas_id,omitempty"`
	FakultasNama    string         `json:"fakultas_nama,omitempty"`
	ProdiID         string         `json:"prodi_id,omitempty"`
	ProdiNama       string         `json:"prodi_nama,omitempty"`
	DosenPembimbing string         `json:"dosen_pembimbing,omitempty"`
	FilePath        string         `json:"file_path,omitempty"`
	Status          string         `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	Files           []DocumentFile `json:"files,omitempty"`
}

// DocumentFile mewakili file yang terkait dengan dokumen
type DocumentFile struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	FileSize   int64  `json:"file_size"`
	FileOrder  int    `json:"file_order"`
	IsLocked   bool   `json:"is_locked"`
}

// CreateDocumentRequest adalah request body untuk membuat dokumen baru
type CreateDocumentRequest struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	Category        string `json:"category"`
	Status          string `json:"status"`
	FakultasID      string `json:"fakultas_id"`
	ProdiID         string `json:"prodi_id"`
	DosenPembimbing string `json:"dosen_pembimbing"`
}

// UpdateDocumentRequest adalah request body untuk update dokumen
type UpdateDocumentRequest struct {
	Title           string `json:"title"`
	Author          string `json:"author"`
	Category        string `json:"category"`
	Status          string `json:"status"`
	FakultasID      string `json:"fakultas_id"`
	ProdiID         string `json:"prodi_id"`
	DosenPembimbing string `json:"dosen_pembimbing"`
}
