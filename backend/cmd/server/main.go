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
	"net"
	"net/http"
	"os"
	"strings"

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

	// --- Email OTP Verification (Public) ---
	// Verifikasi email via OTP sebelum submit access request
	http.HandleFunc("/api/send-otp", handlers.SendOTPHandler)
	http.HandleFunc("/api/verify-otp", handlers.VerifyOTPHandler)

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

	// --- Student Registration Routes ---
	// POST: Public (mahasiswa bisa mendaftar tanpa login)
	// GET/PUT/DELETE: Admin only (mengelola pendaftaran)
	http.HandleFunc("/api/student-signup", handlers.StudentSignupHandler)
	http.HandleFunc("/api/student-registrations", middleware.AdminMiddleware(handlers.StudentRegistrationsHandler))
	http.HandleFunc("/api/student-registrations/", middleware.AdminMiddleware(handlers.StudentRegistrationByIdHandler))

	// --- Serve Frontend Static Files ---
	// Jika folder ../frontend/dist ada, serve SPA dari situ
	frontendDir := "../frontend/dist"
	if _, err := os.Stat(frontendDir); err == nil {
		log.Println("✅ Frontend build ditemukan, serving dari", frontendDir)

		fs := http.FileServer(http.Dir(frontendDir))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Jika request ke /api/ sudah di-handle di atas, ini catch-all
			path := r.URL.Path

			// Skip API routes dan uploads (sudah di-handle)
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/uploads/") || strings.HasPrefix(path, "/download/") {
				http.NotFound(w, r)
				return
			}

			// Coba serve file statis (css, js, gambar, dll)
			filePath := frontendDir + path
			if _, err := os.Stat(filePath); err == nil && path != "/" {
				fs.ServeHTTP(w, r)
				return
			}

			// Fallback: serve index.html untuk SPA routing
			http.ServeFile(w, r, frontendDir+"/index.html")
		})
	} else {
		log.Println("⚠️  Frontend build tidak ditemukan. Jalankan 'npm run build' di folder frontend/")
		log.Println("   Server hanya berjalan sebagai API backend.")
	}

	// ============================================
	// START SERVER
	// ============================================
	port := ":8080"
	localIP := getLocalIP()

	fmt.Println("========================================")
	fmt.Println("  Repository UN - Server")
	fmt.Println("========================================")
	fmt.Println("")
	fmt.Printf("  ➜  Local:   http://localhost%s\n", port)
	if localIP != "" {
		fmt.Printf("  ➜  Network: http://%s%s\n", localIP, port)
	}
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

	// Listen di 0.0.0.0 agar bisa diakses dari perangkat lain di jaringan
	if err := http.ListenAndServe("0.0.0.0"+port, nil); err != nil {
		log.Fatal("❌ Server gagal start: ", err)
	}
}

// getLocalIP mendeteksi IP lokal laptop di jaringan
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
