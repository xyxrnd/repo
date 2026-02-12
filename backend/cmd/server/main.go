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
	"net/http"

	"repository-un/internal/config"
	"repository-un/internal/handlers"
	"repository-un/internal/middleware"
)

func main() {
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

	// --- User Routes (Admin Only) ---
	// Hanya admin yang bisa mengelola user
	http.HandleFunc("/api/users", middleware.AdminMiddleware(handlers.UsersHandler))
	http.HandleFunc("/api/users/", middleware.AdminMiddleware(handlers.UserByIdHandler))

	// --- Document Routes ---
	// CRUD dokumen
	http.HandleFunc("/uploads", handlers.UploadHandler)
	http.HandleFunc("/api/documents", handlers.DocumentsHandler)
	http.HandleFunc("/api/documents/", handlers.DocumentByIdHandler)

	// --- File Routes ---
	// Download file
	http.HandleFunc("/download/", handlers.DownloadHandler)
	// Serve uploaded files statically (for individual file downloads)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// --- Fakultas Routes (Admin Only) ---
	http.HandleFunc("/api/fakultas", middleware.AdminMiddleware(handlers.FakultasHandler))
	http.HandleFunc("/api/fakultas/", middleware.AdminMiddleware(handlers.FakultasByIdHandler))

	// --- Prodi Routes (Admin Only) ---
	http.HandleFunc("/api/prodi", middleware.AdminMiddleware(handlers.ProdiHandler))
	http.HandleFunc("/api/prodi/", middleware.AdminMiddleware(handlers.ProdiByIdHandler))

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

	http.ListenAndServe(":8080", nil)
}
