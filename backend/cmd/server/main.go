/*
Repository UN - Backend Server
==============================
Server API untuk aplikasi repository dokumen.

Struktur folder:
  - cmd/server/     : Entry point aplikasi
  - internal/       : Kode internal aplikasi
    - config/       : Konfigurasi (database, dll)
    - handlers/     : HTTP handlers untuk setiap endpoint
    - middleware/   : Middleware (auth, cors, dll)
    - models/       : Struktur data (User, Document, dll)
    - utils/        : Fungsi utilitas (PDF processing, dll)
  - migrations/     : SQL migration files
  - uploads/        : File yang diupload

Cara menjalankan:
  go run cmd/server/main.go

Atau build dan run:
  go build -o server.exe cmd/server/main.go
  ./server.exe
*/

package main

import (
	"fmt"
	"log"
	"net/http"

	"repository-un/internal/config"
	"repository-un/internal/handlers"
	"repository-un/internal/middleware"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (jika ada)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  File .env tidak ditemukan, menggunakan environment variables dari OS")
	} else {
		log.Println("✅ Konfigurasi .env berhasil dimuat")
	}

	// Koneksi ke database
	config.ConnectDB()

	// ============================================
	// ROUTES
	// ============================================

	// --- Auth Routes (Public) ---
	// Login dan register tidak perlu token
	http.HandleFunc("/api/auth/login", handlers.LoginHandler)
	http.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	http.HandleFunc("/api/auth/me", middleware.AuthMiddleware(handlers.GetMeHandler))

	// --- Access Request Routes ---
	// POST: Public (siapa saja bisa request akses)
	// GET: Admin (list semua access requests)
	http.HandleFunc("/api/access-requests", handlers.AccessRequestHandler)
	http.HandleFunc("/api/access-requests/", middleware.AdminMiddleware(handlers.AccessRequestByIdHandler))

	// --- Verify Access Token (Public) ---
	http.HandleFunc("/api/verify-access-token", handlers.VerifyAccessTokenHandler)

	// --- User Routes (Admin Only) ---
	// Hanya admin yang bisa mengelola user
	http.HandleFunc("/api/users", middleware.AdminMiddleware(handlers.UsersHandler))
	http.HandleFunc("/api/users/", middleware.AdminMiddleware(handlers.UserByIdHandler))

	// --- Document Routes ---
	// CRUD dokumen
	http.HandleFunc("/uploads", handlers.UploadHandler)
	http.HandleFunc("/api/documents", handlers.DocumentsHandler)
	http.HandleFunc("/api/documents/popular", handlers.PopularDocumentsHandler) // Must be before /api/documents/
	http.HandleFunc("/api/documents/", handlers.DocumentByIdHandler)

	// --- File Routes ---
	// Download file
	http.HandleFunc("/download/", handlers.DownloadHandler)
	// Serve uploaded files statically (for individual file downloads)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// --- Fakultas Routes ---
	// GET: Public (untuk filter dropdown)
	// POST: Admin only
	http.HandleFunc("/api/fakultas", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodOptions {
			handlers.FakultasHandler(w, r)
		} else {
			middleware.AdminMiddleware(handlers.FakultasHandler)(w, r)
		}
	})
	http.HandleFunc("/api/fakultas/", middleware.AdminMiddleware(handlers.FakultasByIdHandler))

	// --- Prodi Routes ---
	// GET: Public (untuk filter dropdown)
	// POST: Admin only
	http.HandleFunc("/api/prodi", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodOptions {
			handlers.ProdiHandler(w, r)
		} else {
			middleware.AdminMiddleware(handlers.ProdiHandler)(w, r)
		}
	})
	http.HandleFunc("/api/prodi/", middleware.AdminMiddleware(handlers.ProdiByIdHandler))

	// --- Site Settings Routes ---
	// GET: Public (untuk ambil nama app, logo, dll)
	// PUT: Admin only (untuk update settings)
	http.HandleFunc("/api/site-settings", handlers.SiteSettingsHandler)
	http.HandleFunc("/api/site-settings/logo", middleware.AdminMiddleware(handlers.SiteLogoHandler))

	// ============================================
	// START SERVER
	// ============================================
	fmt.Println("========================================")
	fmt.Println("  Repository UN - Backend Server")
	fmt.Println("========================================")
	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("")
	fmt.Println("Endpoints:")
	fmt.Println("  POST /api/auth/login     - Login")
	fmt.Println("  POST /api/auth/register  - Register")
	fmt.Println("  GET  /api/auth/me        - Get current user")
	fmt.Println("  GET  /api/users          - List users (admin)")
	fmt.Println("  GET  /api/documents      - List documents")
	fmt.Println("  POST /api/documents      - Create document")
	fmt.Println("  GET  /api/fakultas       - List fakultas (admin)")
	fmt.Println("  GET  /api/prodi          - List prodi (admin)")
	fmt.Println("")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("❌ Server gagal start: ", err)
	}
}
