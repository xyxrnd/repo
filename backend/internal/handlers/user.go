package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"repository-un/internal/config"
	"repository-un/internal/middleware"
	"repository-un/internal/models"
	"repository-un/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UsersHandler menangani operasi list dan create user
// GET /api/users - List semua user
// POST /api/users - Buat user baru
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	switch r.Method {
	case http.MethodGet:
		listUsers(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// UserByIdHandler menangani operasi pada user tertentu
// GET /api/users/:id - Get user by ID
// PUT /api/users/:id - Update user
// DELETE /api/users/:id - Delete user
func UserByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	id := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if id == "" {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getUserById(w, r, id)
	case http.MethodPut:
		updateUser(w, r, id)
	case http.MethodDelete:
		deleteUser(w, r, id)
	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// listUsers mengambil semua user dari database
func listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(context.Background(),
		`SELECT id, name, email, role, created_at, updated_at 
		 FROM users ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch users"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []models.UserResponse{}

	for rows.Next() {
		var u models.UserResponse
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			http.Error(w, `{"error":"Failed to read user data"}`, http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// getUserById mengambil user berdasarkan ID
func getUserById(w http.ResponseWriter, r *http.Request, id string) {
	var u models.UserResponse
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, name, email, role, created_at, updated_at 
		 FROM users WHERE id = $1`, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// createUser membuat user baru (admin only)
func createUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"Name, email, and password are required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}

	if req.Role != "admin" && req.Role != "user" && req.Role != "mahasiswa" {
		http.Error(w, `{"error":"Role must be 'admin', 'user', or 'mahasiswa'"}`, http.StatusBadRequest)
		return
	}

	// Cek apakah email sudah terdaftar
	var exists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, req.Email).Scan(&exists)

	if exists {
		http.Error(w, `{"error":"Email already registered"}`, http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"Failed to process password"}`, http.StatusInternalServerError)
		return
	}

	id := uuid.New().String()
	now := time.Now()

	_, err = config.DB.Exec(context.Background(),
		`INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, req.Name, req.Email, string(hashedPassword), req.Role, now, now,
	)

	if err != nil {
		http.Error(w, `{"error":"Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	response := models.UserResponse{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// updateUser mengupdate data user
func updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, `{"error":"Name and email are required"}`, http.StatusBadRequest)
		return
	}

	if req.Role != "" && req.Role != "admin" && req.Role != "user" && req.Role != "mahasiswa" {
		http.Error(w, `{"error":"Role must be 'admin', 'user', or 'mahasiswa'"}`, http.StatusBadRequest)
		return
	}

	// Cek apakah user ada
	var exists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists)

	if !exists {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	// Cek apakah email sudah dipakai user lain
	var emailExists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)`, req.Email, id).Scan(&emailExists)

	if emailExists {
		http.Error(w, `{"error":"Email already in use"}`, http.StatusConflict)
		return
	}

	now := time.Now()

	if req.Password != "" {
		if len(req.Password) < 6 {
			http.Error(w, `{"error":"Password must be at least 6 characters"}`, http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error":"Failed to process password"}`, http.StatusInternalServerError)
			return
		}

		_, err = config.DB.Exec(context.Background(),
			`UPDATE users SET name = $1, email = $2, password = $3, role = $4, updated_at = $5 WHERE id = $6`,
			req.Name, req.Email, string(hashedPassword), req.Role, now, id,
		)
		if err != nil {
			http.Error(w, `{"error":"Failed to update user"}`, http.StatusInternalServerError)
			return
		}
	} else {
		_, err := config.DB.Exec(context.Background(),
			`UPDATE users SET name = $1, email = $2, role = $3, updated_at = $4 WHERE id = $5`,
			req.Name, req.Email, req.Role, now, id,
		)
		if err != nil {
			http.Error(w, `{"error":"Failed to update user"}`, http.StatusInternalServerError)
			return
		}
	}

	// Ambil data user yang sudah diupdate
	var u models.UserResponse
	config.DB.QueryRow(context.Background(),
		`SELECT id, name, email, role, created_at, updated_at FROM users WHERE id = $1`, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// deleteUser menghapus user
func deleteUser(w http.ResponseWriter, r *http.Request, id string) {
	// Tidak boleh menghapus diri sendiri
	currentUserID := r.Header.Get("X-User-ID")
	if currentUserID == id {
		http.Error(w, `{"error":"Cannot delete your own account"}`, http.StatusForbidden)
		return
	}

	// Ambil email user untuk mencari data registrasi & KTM
	var userEmail string
	config.DB.QueryRow(context.Background(),
		`SELECT email FROM users WHERE id = $1`, id).Scan(&userEmail)

	// Hapus file KTM dari Google Drive jika ada data registrasi
	if userEmail != "" {
		var ktmPath string
		config.DB.QueryRow(context.Background(),
			`SELECT COALESCE(ktm_path, '') FROM student_registrations WHERE email = $1`, userEmail).Scan(&ktmPath)
		if ktmPath != "" && utils.IsGDriveID(ktmPath) {
			if err := utils.DeleteFromGDrive(ktmPath); err != nil {
				fmt.Printf("⚠️ Gagal hapus KTM dari Google Drive: %v\n", err)
			} else {
				fmt.Printf("✅ KTM berhasil dihapus dari Google Drive (user: %s)\n", userEmail)
			}
		}

		// Hapus data registrasi mahasiswa
		config.DB.Exec(context.Background(),
			`DELETE FROM student_registrations WHERE email = $1`, userEmail)
	}

	result, err := config.DB.Exec(context.Background(),
		`DELETE FROM users WHERE id = $1`, id)

	if err != nil {
		http.Error(w, `{"error":"Failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"User deleted successfully"}`))
}
