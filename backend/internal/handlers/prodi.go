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

// ProdiHandler menangani operasi list dan create prodi
// GET /api/prodi - List semua prodi
// POST /api/prodi - Buat prodi baru
func ProdiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	switch r.Method {
	case http.MethodGet:
		listProdi(w, r)
	case http.MethodPost:
		createProdi(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ProdiByIdHandler menangani operasi pada prodi tertentu
// GET /api/prodi/:id - Get prodi by ID
// PUT /api/prodi/:id - Update prodi
// DELETE /api/prodi/:id - Delete prodi
func ProdiByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	id := strings.TrimPrefix(r.URL.Path, "/api/prodi/")
	if id == "" {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getProdiById(w, r, id)
	case http.MethodPut:
		updateProdi(w, r, id)
	case http.MethodDelete:
		deleteProdi(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listProdi mengambil semua prodi dari database dengan nama fakultas
func listProdi(w http.ResponseWriter, r *http.Request) {
	// Support filter by fakultas_id
	fakultasID := r.URL.Query().Get("fakultas_id")

	var query string
	var args []interface{}

	if fakultasID != "" {
		query = `SELECT p.id, p.nama, p.kode, p.fakultas_id, f.nama as fakultas_nama, p.created_at, p.updated_at
			 FROM prodi p
			 LEFT JOIN fakultas f ON p.fakultas_id = f.id
			 WHERE p.fakultas_id = $1
			 ORDER BY p.nama ASC`
		args = append(args, fakultasID)
	} else {
		query = `SELECT p.id, p.nama, p.kode, p.fakultas_id, f.nama as fakultas_nama, p.created_at, p.updated_at
			 FROM prodi p
			 LEFT JOIN fakultas f ON p.fakultas_id = f.id
			 ORDER BY f.nama ASC, p.nama ASC`
	}

	rows, err := config.DB.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, "Gagal mengambil data prodi", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	prodiList := []models.Prodi{}

	for rows.Next() {
		var p models.Prodi
		err := rows.Scan(
			&p.ID,
			&p.Nama,
			&p.Kode,
			&p.FakultasID,
			&p.FakultasNama,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Gagal membaca data", http.StatusInternalServerError)
			return
		}
		prodiList = append(prodiList, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prodiList)
}

// getProdiById mengambil prodi berdasarkan ID
func getProdiById(w http.ResponseWriter, r *http.Request, id string) {
	var p models.Prodi
	err := config.DB.QueryRow(context.Background(),
		`SELECT p.id, p.nama, p.kode, p.fakultas_id, f.nama as fakultas_nama, p.created_at, p.updated_at
		 FROM prodi p
		 LEFT JOIN fakultas f ON p.fakultas_id = f.id
		 WHERE p.id = $1`, id).Scan(
		&p.ID,
		&p.Nama,
		&p.Kode,
		&p.FakultasID,
		&p.FakultasNama,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Program Studi tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// createProdi membuat prodi baru
func createProdi(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProdiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request body tidak valid", http.StatusBadRequest)
		return
	}

	if req.Nama == "" || req.Kode == "" || req.FakultasID == "" {
		http.Error(w, "Nama, Kode, dan Fakultas wajib diisi", http.StatusBadRequest)
		return
	}

	// Validasi fakultas exists
	var fakExists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM fakultas WHERE id = $1)`, req.FakultasID).Scan(&fakExists)
	if !fakExists {
		http.Error(w, "Fakultas tidak ditemukan", http.StatusBadRequest)
		return
	}

	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO prodi (id, nama, kode, fakultas_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := config.DB.Exec(context.Background(), query,
		id, req.Nama, req.Kode, req.FakultasID, now, now,
	)

	if err != nil {
		http.Error(w, "Gagal menyimpan prodi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get fakultas nama for response
	var fakNama string
	config.DB.QueryRow(context.Background(),
		`SELECT nama FROM fakultas WHERE id = $1`, req.FakultasID).Scan(&fakNama)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.Prodi{
		ID:           id.String(),
		Nama:         req.Nama,
		Kode:         req.Kode,
		FakultasID:   req.FakultasID,
		FakultasNama: fakNama,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

// updateProdi mengupdate data prodi
func updateProdi(w http.ResponseWriter, r *http.Request, id string) {
	var req models.UpdateProdiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request body tidak valid", http.StatusBadRequest)
		return
	}

	if req.Nama == "" || req.Kode == "" || req.FakultasID == "" {
		http.Error(w, "Nama, Kode, dan Fakultas wajib diisi", http.StatusBadRequest)
		return
	}

	// Validasi fakultas exists
	var fakExists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM fakultas WHERE id = $1)`, req.FakultasID).Scan(&fakExists)
	if !fakExists {
		http.Error(w, "Fakultas tidak ditemukan", http.StatusBadRequest)
		return
	}

	now := time.Now()

	query := `
		UPDATE prodi
		SET nama = $1, kode = $2, fakultas_id = $3, updated_at = $4
		WHERE id = $5
	`
	result, err := config.DB.Exec(context.Background(), query,
		req.Nama, req.Kode, req.FakultasID, now, id)

	if err != nil {
		http.Error(w, "Gagal update prodi", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Program Studi tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          id,
		"nama":        req.Nama,
		"kode":        req.Kode,
		"fakultas_id": req.FakultasID,
	})
}

// deleteProdi menghapus prodi
func deleteProdi(w http.ResponseWriter, r *http.Request, id string) {
	result, err := config.DB.Exec(
		context.Background(),
		`DELETE FROM prodi WHERE id = $1`,
		id,
	)
	if err != nil {
		http.Error(w, "Gagal menghapus program studi", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, "Program Studi tidak ditemukan", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Program Studi berhasil dihapus"}`))
}
