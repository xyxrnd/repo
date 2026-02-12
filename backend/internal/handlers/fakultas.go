package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"repository-un/internal/config"
	"repository-un/internal/middleware"
	"repository-un/internal/models"

	"github.com/google/uuid"
)

// FakultasHandler menangani operasi list dan create fakultas
// GET /api/fakultas - List semua fakultas
// POST /api/fakultas - Buat fakultas baru
func FakultasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	switch r.Method {
	case http.MethodGet:
		listFakultas(w, r)
	case http.MethodPost:
		createFakultas(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// FakultasByIdHandler menangani operasi pada fakultas tertentu
// GET /api/fakultas/:id - Get fakultas by ID
// PUT /api/fakultas/:id - Update fakultas
// DELETE /api/fakultas/:id - Delete fakultas
func FakultasByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	id := strings.TrimPrefix(r.URL.Path, "/api/fakultas/")
	if id == "" {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getFakultasById(w, r, id)
	case http.MethodPut:
		updateFakultas(w, r, id)
	case http.MethodDelete:
		deleteFakultas(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listFakultas mengambil semua fakultas dari database
func listFakultas(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(context.Background(),
		`SELECT id, nama, kode, created_at, updated_at
		 FROM fakultas
		 ORDER BY nama ASC`)
	if err != nil {
		http.Error(w, "Gagal mengambil data fakultas", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fakultasList := []models.Fakultas{}

	for rows.Next() {
		var f models.Fakultas
		err := rows.Scan(
			&f.ID,
			&f.Nama,
			&f.Kode,
			&f.CreatedAt,
			&f.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Gagal membaca data", http.StatusInternalServerError)
			return
		}
		fakultasList = append(fakultasList, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fakultasList)
}

// getFakultasById mengambil fakultas berdasarkan ID
func getFakultasById(w http.ResponseWriter, r *http.Request, id string) {
	var f models.Fakultas
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, nama, kode, created_at, updated_at
		 FROM fakultas WHERE id = $1`, id).Scan(
		&f.ID,
		&f.Nama,
		&f.Kode,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Fakultas tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// createFakultas membuat fakultas baru
func createFakultas(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFakultasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request body tidak valid", http.StatusBadRequest)
		return
	}

	if req.Nama == "" || req.Kode == "" {
		http.Error(w, "Nama dan Kode wajib diisi", http.StatusBadRequest)
		return
	}

	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO fakultas (id, nama, kode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := config.DB.Exec(context.Background(), query,
		id, req.Nama, req.Kode, now, now,
	)

	if err != nil {
		http.Error(w, "Gagal menyimpan fakultas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.Fakultas{
		ID:        id.String(),
		Nama:      req.Nama,
		Kode:      req.Kode,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// updateFakultas mengupdate data fakultas
func updateFakultas(w http.ResponseWriter, r *http.Request, id string) {
	var req models.UpdateFakultasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request body tidak valid", http.StatusBadRequest)
		return
	}

	if req.Nama == "" || req.Kode == "" {
		http.Error(w, "Nama dan Kode wajib diisi", http.StatusBadRequest)
		return
	}

	now := time.Now()

	query := `
		UPDATE fakultas
		SET nama = $1, kode = $2, updated_at = $3
		WHERE id = $4
	`
	result, err := config.DB.Exec(context.Background(), query,
		req.Nama, req.Kode, now, id)

	if err != nil {
		http.Error(w, "Gagal update fakultas", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Fakultas tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   id,
		"nama": req.Nama,
		"kode": req.Kode,
	})
}

// deleteFakultas menghapus fakultas
func deleteFakultas(w http.ResponseWriter, r *http.Request, id string) {
	// Cek apakah ada prodi yang terkait
	var prodiCount int
	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM prodi WHERE fakultas_id = $1`, id).Scan(&prodiCount)

	if prodiCount > 0 {
		http.Error(w, "Tidak dapat menghapus fakultas yang masih memiliki program studi", http.StatusBadRequest)
		return
	}

	result, err := config.DB.Exec(
		context.Background(),
		`DELETE FROM fakultas WHERE id = $1`,
		id,
	)
	if err != nil {
		http.Error(w, "Gagal menghapus fakultas", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Fakultas tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Fakultas berhasil dihapus"}`))
}
