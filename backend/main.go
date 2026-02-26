/*
Repository UN - Backend Server
==============================

PERHATIAN: File ini adalah entry point lama.
Gunakan: go run cmd/server/main.go

File ini tetap dipertahankan untuk backward compatibility.
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

	// --- Auth Routes (Public) ---
	http.HandleFunc("/api/auth/login", handlers.LoginHandler)
	http.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	http.HandleFunc("/api/auth/me", middleware.AuthMiddleware(handlers.GetMeHandler))

	// --- User Routes (Admin Only) ---
	http.HandleFunc("/api/users", middleware.AdminMiddleware(handlers.UsersHandler))
	http.HandleFunc("/api/users/", middleware.AdminMiddleware(handlers.UserByIdHandler))

	// --- Document Routes ---
	http.HandleFunc("/uploads", handlers.UploadHandler)
	http.HandleFunc("/api/documents", handlers.DocumentsHandler)
	http.HandleFunc("/api/documents/popular", handlers.PopularDocumentsHandler)
	http.HandleFunc("/api/documents/", handlers.DocumentByIdHandler)

	// --- File Routes ---
	http.HandleFunc("/download/", handlers.DownloadHandler)
	// Serve uploaded files statically (for individual file downloads)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// --- Fakultas Routes (Admin Only) ---
	http.HandleFunc("/api/fakultas", middleware.AdminMiddleware(handlers.FakultasHandler))
	http.HandleFunc("/api/fakultas/", middleware.AdminMiddleware(handlers.FakultasByIdHandler))

	// --- Prodi Routes (Admin Only) ---
	http.HandleFunc("/api/prodi", middleware.AdminMiddleware(handlers.ProdiHandler))
	http.HandleFunc("/api/prodi/", middleware.AdminMiddleware(handlers.ProdiByIdHandler))

	// --- Site Settings Routes ---
	http.HandleFunc("/api/site-settings", handlers.SiteSettingsHandler)
	http.HandleFunc("/api/site-settings/logo", middleware.AdminMiddleware(handlers.SiteLogoHandler))

	// --- Student Registration Routes ---
	http.HandleFunc("/api/student-signup", handlers.StudentSignupHandler) // Public
	http.HandleFunc("/api/student-registrations", middleware.AdminMiddleware(handlers.StudentRegistrationsHandler))
	http.HandleFunc("/api/student-registrations/", middleware.AdminMiddleware(handlers.StudentRegistrationByIdHandler))

	fmt.Println("========================================")
	fmt.Println("  Repository UN - Backend Server")
	fmt.Println("========================================")
	fmt.Println("Server running at http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
