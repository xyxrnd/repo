package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"repository-un/internal/config"
	"repository-un/internal/middleware"

	"github.com/google/uuid"
)

// SiteSettingsHandler mengelola endpoint /api/site-settings
func SiteSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	switch r.Method {
	case http.MethodGet:
		getSiteSettings(w, r)
	case http.MethodPut:
		updateSiteSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// SiteLogoHandler mengelola upload logo di /api/site-settings/logo
func SiteLogoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		middleware.EnableCORS(w)
		return
	}
	middleware.EnableCORS(w)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uploadSiteLogo(w, r)
}

// getSiteSettings mengambil semua pengaturan situs
func getSiteSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query(context.Background(),
		`SELECT key, value FROM site_settings`)
	if err != nil {
		http.Error(w, "Gagal mengambil pengaturan: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		settings[key] = value
	}

	// Return structured response with defaults
	response := map[string]string{
		"app_name":        getSettingOrDefault(settings, "app_name", "ScholarHub"),
		"app_description": getSettingOrDefault(settings, "app_description", "Repository Dokumen Akademik"),
		"about_text":      getSettingOrDefault(settings, "about_text", "Platform digital untuk menyimpan, mengelola, dan menyebarluaskan karya ilmiah civitas akademika."),
		"visi":            getSettingOrDefault(settings, "visi", ""),
		"misi":            getSettingOrDefault(settings, "misi", ""),
		"logo_url":        getSettingOrDefault(settings, "logo_url", ""),
		"address":         getSettingOrDefault(settings, "address", "Jl. Universitas No. 1, Indonesia"),
		"email":           getSettingOrDefault(settings, "email", "repository@univ.ac.id"),
		"phone":           getSettingOrDefault(settings, "phone", "(021) 1234567"),
		"footer_text":     getSettingOrDefault(settings, "footer_text", ""),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getSettingOrDefault(settings map[string]string, key string, defaultValue string) string {
	if val, ok := settings[key]; ok && val != "" {
		return val
	}
	return defaultValue
}

// updateSiteSettings memperbarui pengaturan situs
func updateSiteSettings(w http.ResponseWriter, r *http.Request) {
	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Format data tidak valid", http.StatusBadRequest)
		return
	}

	// Allowed keys
	allowedKeys := map[string]bool{
		"app_name":        true,
		"app_description": true,
		"about_text":      true,
		"visi":            true,
		"misi":            true,
		"logo_url":        true,
		"address":         true,
		"email":           true,
		"phone":           true,
		"footer_text":     true,
	}

	for key, value := range updates {
		if !allowedKeys[key] {
			continue
		}

		_, err := config.DB.Exec(context.Background(),
			`INSERT INTO site_settings (key, value, updated_at) 
			 VALUES ($1, $2, CURRENT_TIMESTAMP)
			 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP`,
			key, value)
		if err != nil {
			http.Error(w, fmt.Sprintf("Gagal menyimpan pengaturan '%s': %s", key, err.Error()), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Pengaturan berhasil disimpan",
		"updated": updates,
	})
}

// uploadSiteLogo mengupload logo situs
func uploadSiteLogo(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB max

	file, header, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, "File logo tidak ditemukan", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validasi file type
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".svg": true, ".webp": true, ".ico": true}
	if !allowedExts[ext] {
		http.Error(w, "Format file tidak didukung. Gunakan PNG, JPG, SVG, WEBP, atau ICO.", http.StatusBadRequest)
		return
	}

	// Pastikan folder uploads ada
	os.MkdirAll("uploads/logo", os.ModePerm)

	// Simpan file
	storedName := "site_logo_" + uuid.New().String() + ext
	filePath := "uploads/logo/" + storedName

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Gagal menyimpan file logo", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Simpan path ke database
	logoURL := "/uploads/logo/" + storedName
	_, err = config.DB.Exec(context.Background(),
		`INSERT INTO site_settings (key, value, updated_at) 
		 VALUES ('logo_url', $1, CURRENT_TIMESTAMP)
		 ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = CURRENT_TIMESTAMP`,
		logoURL)
	if err != nil {
		http.Error(w, "Gagal menyimpan path logo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Logo berhasil diupload",
		"logo_url": logoURL,
	})
}
