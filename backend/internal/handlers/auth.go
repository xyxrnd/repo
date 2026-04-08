package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"repository-un/internal/config"
	"repository-un/internal/middleware"
	"repository-un/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler menangani proses login user
// POST /api/auth/login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"Email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Ambil user dari database
	var user models.User
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, name, email, password, role, created_at, updated_at 
		 FROM users WHERE email = $1`, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User: models.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterHandler menangani proses registrasi user baru
// POST /api/auth/register
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
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

	// Insert user dengan role default "admin"
	_, err = config.DB.Exec(context.Background(),
		`INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'admin', $5, $6)`,
		id, req.Name, req.Email, string(hashedPassword), now, now,
	)

	if err != nil {
		http.Error(w, `{"error":"Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(id, req.Email, "admin")
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User: models.UserResponse{
			ID:        id,
			Name:      req.Name,
			Email:     req.Email,
			Role:      "admin",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetMeHandler mengembalikan data user yang sedang login
// GET /api/auth/me
func GetMeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Ambil user ID dari header (di-set oleh AuthMiddleware)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, `{"error":"User not found"}`, http.StatusUnauthorized)
		return
	}

	var user models.UserResponse
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, name, email, role, created_at, updated_at 
		 FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
